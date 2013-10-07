package main

func guard(jsonConfig jsonConfig) {
	for _, json := range jsonConfig {
		logger.Println(json)

	}

}
