var builder = DistributedApplication.CreateBuilder(args);

// --------------------
// Backend (Go + air)
// --------------------
var backend = builder
    .AddExecutable(name: "backend", command: "air", workingDirectory: "../backend", args: [])
    .WithEnvironment("FRONTEND_MODE", "dev")
    .WithHttpEndpoint(env: "8080");
;

// --------------------
// Frontend (Svelte / Vite)
// --------------------
var frontend = builder.AddExecutable(
    name: "frontend",
    command: "npm",
    workingDirectory: "../frontend",
    args: ["run", "dev", "--", "--port", "5173", "--strictPort"]
);

// --------------------
// MPV player
// --------------------
var mpv = builder.AddExecutable(
    name: "mpv",
    command: "mpv",
    workingDirectory: ".",
    args: ["--idle=yes", "--input-ipc-server=/tmp/mpv.sock", "--force-window=yes"]
);

// --------------------
// Start everything
// --------------------
builder.Build().Run();
