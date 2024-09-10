---@diagnostic disable: duplicate-set-field
-- _embedded_loader is a function registered in go
table.insert(package.loaders, 2, gsharer._embedded_loader)
gsharer.json = require("gsharer.lib.json")
gsharer.inspect = require("gsharer.lib.inspect")

function gsharer.print(...)
	for i = 1, select("#", ...) do
		local o = select(i, ...)
		if type(o) == "string" then
			print(o)
		else
			print(gsharer.inspect(o, { newline = "\n", indent = "  " }))
		end
		print("\n")
	end

	return ...
end

---@return gsharer.Destination
function gsharer.auto(file_info)
	return require("gsharer.destinations.litterbox.file")
end

function gsharer.option(name, default)
	if os.getenv("GSHARER_INTERACTIVE") and os.getenv("GSHARER_INTERACTIVE") ~= "0" then
		io.write(string.format("Input the value for option %s (default %s): ", name, default))
		local line = io.read()
		return line
	else
		return os.getenv(name) or _G[name]
	end
end

function gsharer.gsplit(s, delimiter)
	return (s .. delimiter):gmatch("(.-)" .. delimiter)
end

function gsharer.split(s, delimiter)
	local result = {}
	for match in gsharer.gsplit(s, delimiter) do
		table.insert(result, match)
	end
	return result
end
