local M = {}
-- {
--   "Version": "14.1.0",
--   "DestinationType": "ImageUploader",
--   "RequestMethod": "POST",
--   "RequestURL": "https://lensdump.com/api/1/upload",
--   "Body": "MultipartFormData",
--   "Arguments": {
--     "key": "Replace with your API KEY from https://lensdump.com/settings/api",
--     "album_id": "Replace with Album ID where you want to upload images (optional)",
--     "source": "{input}"
--   },
--   "FileFormName": "source",
--   "URL": "{json:image.url}",
--   "ThumbnailURL": "{json:thumb.url}",
--   "DeletionURL": "{json:delete_url}",
--   "ErrorMessage": "{json:status_txt}"
-- }
local function convert_field(str) end
function M.convert(sxcu, name)
	local json = gsharer.json.decode(sxcu)
	local destination = {
		name = name,
		request = {
			method = json["RequestMethod"],
			URL = json["RequestURL"],
			arguments = json["Arguments"],
			file_form_name = json["FileFormName"],
		},
		response = function(response_str) end,
	}
	destination.request.arguments = json["Arguments"]
	return lua_table
end
return M
