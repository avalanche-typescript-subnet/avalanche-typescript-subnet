package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/hypersdk/pebble"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/ava-labs/hypersdk/x/programs/examples"
	"github.com/ava-labs/hypersdk/x/programs/examples/simulator/cmd"
	"github.com/ava-labs/hypersdk/x/programs/runtime"
	xutils "github.com/ava-labs/hypersdk/x/programs/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type ProgramSimulator struct {
	log logging.Logger
	db  database.Database
}

func newProgramPublish(log logging.Logger, db database.Database) *ProgramSimulator {
	return &ProgramSimulator{
		log: log,
		db:  db,
	}
}

func main() {
	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"} // Update with your frontend URL

	// set log and db
	log := xutils.NewLoggerWithLogLevel(logging.Debug)
	db, _, err := pebble.New(examples.DBPath, pebble.NewDefaultConfig())
	if err != nil {
		return
	}
	utils.Outf("{{yellow}}database:{{/}} %s\n", examples.DBPath)
	programPublish := newProgramPublish(log, db)
	r.Use(cors.New(config))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/api/keys", programPublish.keysHandler)
	r.POST("/api/publish", programPublish.publishHandler)
	r.POST("/api/invoke", programPublish.invokeHandler)

	r.Run(":8080")
}

func (r ProgramSimulator) keysHandler(c *gin.Context) {
	// get keys
	keys, err := cmd.GetKeys(r.db, 5)
	if err != nil {
		fmt.Println("Error: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
		"keys":    keys,
	})
}

func (r ProgramSimulator) publishHandler(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("data: ")
	num, funcs, err := r.PublishProgram(data)
	fmt.Println("funcs: ", funcs)
	// bytes
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Data received successfully", "function_data": funcs, "id": num})
}

func (r ProgramSimulator) invokeHandler(c *gin.Context) {
	var data map[string]interface{}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Process the data
	name, nameExists := data["name"].(string)
	if !nameExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' field"})
		return
	}

	// params is an array of strings
	params, paramsExists := data["params"].([]interface{})
	if !paramsExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'params' field"})
		return
	}
	value, programIDExists := data["programID"].(string)
	if !programIDExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'programID' field"})
		return
	}
	programID, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// now we have the function name, id and the params, can invoke
	result, gas, err := r.invokeProgram(programID, name, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data received and processed successfully", "result": result, "gas": gas})
}

func (r ProgramSimulator) PublishProgram(programBytes []byte) (uint64, map[string]int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runtime := runtime.New(r.log, runtime.NewMeter(r.log, examples.DefaultMaxFee, examples.CostMap), r.db)
	defer runtime.Stop(ctx)

	fmt.Println("in create")
	programID, err := runtime.Create(ctx, programBytes)
	if err != nil {
		return 0, nil, err
	}
	fmt.Println("programID: ", programID)
	data, err := runtime.GetUserData()
	if err != nil {
		return 0, nil, err
	}

	return programID, data, nil
}

func (r ProgramSimulator) invokeProgram(programID uint64, functionName string, params []interface{}) (uint64, uint64, error) {
	exists, program, err := runtime.GetProgram(r.db, programID)
	if !exists {
		return 0, 0, fmt.Errorf("program %v does not exist", programID)
	}
	if err != nil {
		return 0, 0, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: owner for now, change to caller later
	runtime := runtime.New(r.log, runtime.NewMeter(r.log, examples.DefaultMaxFee, examples.CostMap), r.db)
	defer runtime.Stop(ctx)

	err = runtime.Initialize(ctx, program)
	if err != nil {
		return 0, 0, err
	}

	var callParams []uint64

	if len(params) > 0 {
		for _, param := range params {
			fmt.Printf("Type of x: %T\n", param)
			switch p := strings.ToLower(param.(string)); {
			case p == "true":
				callParams = append(callParams, 1)
			case p == "false":
				callParams = append(callParams, 0)
			case strings.HasPrefix(p, cmd.HRP):
				// address
				pk, err := cmd.GetPublicKey(r.db, p)
				if err != nil {
					return 0, 0, err
				}
				ptr, err := runtime.WriteGuestBuffer(ctx, pk[:])
				if err != nil {
					return 0, 0, err
				}
				callParams = append(callParams, ptr)
			default:
				// treat like a number
				var num uint64
				num, err := strconv.ParseUint(p, 10, 64)
				if err != nil {
					return 0, 0, err
				}
				callParams = append(callParams, num)
			}
		}
	}
	// prepend programID
	callParams = append([]uint64{programID}, callParams...)

	resp, err := runtime.Call(ctx, functionName, callParams...)
	if err != nil {
		return 0, 0, err
	}

	return resp[0], runtime.GetCurrentGas(ctx), nil
}
