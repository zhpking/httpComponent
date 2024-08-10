package main

func main() {
	server := NewHttpServer("test-server")
	server.Route("/signUp", SignUp)
	server.Start(":8080")
}
