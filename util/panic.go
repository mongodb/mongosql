package util

func PanicSafeGo(function func(), recoverAction func(err interface{})) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				recoverAction(err)
			}
		}()
		function()
	}()
}
