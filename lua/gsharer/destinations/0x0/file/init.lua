---@type gsharer.Destination
return {
	name = "0x0",
	request = {
		method = "POST",
		URL = "https://0x0.st",
		file_form_name = "file",
		arguments = {
			expires = gsharer.option("0x0_EXPIRES"),
		},
	},
	response = function(response_body, headers)
		-- catbox just returns the URL
		gsharer.print(response_body)
		return response_body
	end,
}
