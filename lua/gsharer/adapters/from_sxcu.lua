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
local function convert_to_parser(str) end

---@param sxcu string sxcu file contents (json)
---@param name string the name of the resulting destination
---@return gsharer.Destination
function M.convert(sxcu, name)
	local json = gsharer.json.decode(sxcu)
	---@type gsharer.Destination
	local destination = {
		name = name,
		request = {
			method = json["RequestMethod"],
			URL = json["RequestURL"],
			arguments = json["Arguments"],
			file_form_name = json["FileFormName"],
		},
		response = function(body)
			return body
		end,
	}
	return destination
end
return M
