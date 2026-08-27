package vrpc

import "github.com/vedadiyan/vapor"

func InitServer(server vapor.Server) {
	handlers.Range(func(key, value any) bool {
		fn, ok := value.(RegisterFunc)
		if !ok {
			return true
		}
		if err := server.HandleFunc(fn()); err != nil {
			panic(err)
		}
		return true
	})
}
