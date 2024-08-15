---@meta
gsharer = {}
gsharer.json = {}

---@param str string Stringified JSON data.
---@return any
function gsharer.json.decode(str, opts) end

--- Encodes (or "packs") Lua object {obj} as JSON in a Lua string.
---@param obj any The Lua object to be encoded
---@return string
function gsharer.json.encode(obj) end

--- Creates a human readable string of the lua object given
---@param root any The Lua object to be inspected
---@param options table|nil The options as defined in https://github.com/kikito/inspect.lua/blob/8686162bce74913c4d3a577e7324642ddc4e21c0/inspect.lua#L338
---@return string
function gsharer.inspect(root, options) end
