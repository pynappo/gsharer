return {
	router = function() end,
	auto_filename = function()
		local sets = { { 97, 122 }, { 65, 90 }, { 48, 57 } } -- a-z, A-Z, 0-9
		local function string_random(chars)
			local str = ""
			for i = 1, chars do
				math.randomseed(os.clock() ^ 5)
				local charset = sets[math.random(1, #sets)]
				str = str .. string.char(math.random(charset[1], charset[2]))
			end
			return str
		end
	end,
}
