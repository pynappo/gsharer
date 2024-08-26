-- _embedded_loader is a function registered in go
table.insert(package.loaders, 2, gsharer._embedded_loader)
require("gsharer.lib.json")
require("gsharer.lib.inspect")

--- @see |gsharer.inspect()|
--- @param ... any
--- @return any # given arguments.
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

function gsharer.auto(file_info)
	return require("gsharer.destinations.litterbox.file")
end

--- Gets an option from environment variables or from a lua global.
--- @param name string
--- @return string|nil
function gsharer.option(name)
	local option = os.getenv(name) or _G[name]
	return option
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
