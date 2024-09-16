---@meta
---@class gsharer
gsharer = {}
gsharer.json = {}

--- The loader used to load lua modules that are embedded in the gsharer binary.
---@param modname string Stringified JSON data.
function gsharer._embedded_loader(modname) end

---Decodes JSON into a Lua table.
---@param str string Stringified JSON data.
---@return any
function gsharer.json.decode(str, opts) end

--- Encodes (or "packs") a Lua object as JSON in a Lua string.
---@param obj any The Lua object to be encoded
---@return string
function gsharer.json.encode(obj) end

--- Creates a human readable string of the lua object given
---@param root any The Lua object to be inspected
---@param options table|nil The options as defined in https://github.com/kikito/inspect.lua/blob/8686162bce74913c4d3a577e7324642ddc4e21c0/inspect.lua#L338
---@return string
function gsharer.inspect(root, options) end

--- Prints the lua objects to stdout, separated by newlines.
--- @see gsharer.inspect()
--- @param ... any
--- @return any
function gsharer.print(...) end

---Gets the value of a user-specified option from either the environment variables or lua globals.
---@param name string The name of the option.
---@param default string|nil The default value of the option.
---@return string|nil value The value of the option.
function gsharer.option(name, default) end

---@class gsharer.Request
---@field method "POST"|"GET"|"PUT"
---@field URL string
---@field file_form_name string
---@field arguments {[string]: string|nil}
local Request = {
	method = "POST",
	URL = "https://example.com",
	file_form_name = "fileToUpload",
	arguments = {
		userhash = gsharer.option("DESTINATION_USERHASH"),
		reqtype = "fileupload",
	},
}
---@class gsharer.Destination
---@field name string The name of the destination, for logging purposes
---@field request gsharer.Request The name of the destination, for logging purposes
---@field response fun(body: string, headers: { [string]: string }):string
local Destination = {
	name = "name",
	request = {},
	response = function(body, headers)
		return body
	end,
}
