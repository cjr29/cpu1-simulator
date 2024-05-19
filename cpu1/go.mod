module github.com/cjr29/cpu1-simulator

go 1.21.6

require (
	github.com/cjr29/go6502/asm v0.0.0-20240519135954-448b93fea12a
	github.com/cjr29/go6502/cpu v0.0.0-20240519135954-448b93fea12a
)

replace github.com/cjr29/cpu1-simulator/cpu1 v0.0.0 => ./cpu1

replace github.com/cjr29/cpu1-simulator/dashboard v0.0.0 => ./dashboard
