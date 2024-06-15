1. Build an javy-image 
```bash
docker build -t javy-callback:latest .
```

2. Emit provider
```bash
docker run --rm -v ./:/out javy-callback:latest emit-provider -o /out/javy_provider_1.4.0.wasm
```

3. Move the file ./javy_provider_1.4.0.wasm to the required directory 