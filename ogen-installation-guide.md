# Resolving "ogen" Executable Not Found Error

The error occurs because the required Go tool "ogen" is not installed in your system. Ogen is a Go OpenAPI/Swagger specification generator that needs to be installed globally.

To resolve this error, follow these steps:

1. Make sure you have Go installed on your system (version 1.19 or later recommended)

2. Install the ogen tool using Go:

   ```bash
   go install github.com/ogen-go/ogen/cmd/ogen@latest
   ```

3. Verify the installation:

   ```bash
   ogen --version
   ```

4. Ensure your `$GOPATH/bin` is in your system's `$PATH`. Add this to your shell profile (like `.bashrc`, `.zshrc`, etc.):

   ```bash
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

5. Reload your shell or run:

   ```bash
   source ~/.bashrc  # or source ~/.zshrc
   ```

After completing these steps, the SDK generation should work properly as the `ogen` executable will be available in your system PATH.
