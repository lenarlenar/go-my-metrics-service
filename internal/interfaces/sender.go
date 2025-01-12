package interfaces

type Sender interface {
	Run(reportInterval int, serverAddress string)
}
