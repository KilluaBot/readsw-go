module.exports = {
  apps: [
    {
      name: "golang-app",
      script: "sh",
      args: "-c 'go run main.go'",
      interpreter: "/bin/bash",
      cwd: "~/readsw-go",  // Ganti dengan path proyekmu
      watch: true,  // Optional: untuk otomatis me-restart ketika ada perubahan pada file
    }
  ]
}
