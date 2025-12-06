package websocket_router

// UploadBinary handles binary chunks
// Protocol: [sessionID (36 bytes)][ChunkIndex (4 bytes BigEndian)][Data...]
