return {
	name = "catbox.moe",
	request = {
		type = "POST",
		URL = "https://catbox.moe/user/api.php",
		file_form_name = "fileToUpload",
		arguments = {
			userhash = gsharer.option("CATBOX_USERHASH"),
			reqtype = "fileupload",
		},
		response_type = "text",
	},
	response = function(response_str)
		-- catbox just returns the URL
		gsharer.print(response_str)
		return response_str
	end,
}
