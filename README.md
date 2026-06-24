# OTPKI Protobuf

THIS IS A CHANGE FOR PURPOSE OF THE DEMO

Protobuf definitions and generated code for OTPKI services.

## NPM Package Usage

This repository can be consumed directly as an npm package via git:

```bash
npm install github:otpki/otpki-protobuf
```

### Peer Dependencies

The package requires these peer dependencies:

```bash
npm install @bufbuild/protobuf @connectrpc/connect @connectrpc/connect-query
```

### Imports

```typescript
// Import all generated types
import { ... } from '@otpki/otpki-protobuf';

// Import specific public API types
import { ... } from '@otpki/otpki-protobuf/public';

// Import common types
import { ... } from '@otpki/otpki-protobuf/common';

// Import OpenAPI specs
import spec from '@otpki/otpki-protobuf/specs';
```
