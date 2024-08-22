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
local function convert_from_sxcu(json_str)
	local table = {
		request = {},
		response = function(response_str) end,
	}
	local sxcu_table = gsharer.json.decode(json_str)
	return lua_table
end
