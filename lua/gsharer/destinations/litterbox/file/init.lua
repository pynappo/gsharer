return {
	name = "litterbox.catbox.moe",
	request = {
		type = "POST",
		URL = "https://litterbox.catbox.moe/resources/internals/api.php",
		file_form_name = "fileToUpload",
		arguments = {
			time = gsharer.option("LITTERBOX_TIME") or "1h",
			reqtype = "fileupload",
		},
		response_type = "text",
	},
	response = function(response_str)
		return response_str
	end,
}
