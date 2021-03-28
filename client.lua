local math = require("math")
local mp = require("mp")
local socket = require("socket")
local host, port = "localhost", 27001
local tcp = assert(socket.tcp())

local SEEK_THRESHOLD = 0.4

local EVENT_PLAY = 0
local EVENT_PAUSE = 1
local EVENT_SEEK = 2

local function encode(event, pos)
	return string.format("%d %f\n", event, pos)
end

local function decode(input)
	local t = {}
	for token in string.gmatch(input, "%S+") do
		table.insert(t, token)
	end
	return tonumber(t[1]), tonumber(t[2])
end

local function positionNotNegative(pos)
	pos = pos or mp.get_property("time-pos")
	if pos ~= nil and tonumber(pos) >= 0 then
		return pos
	else
		return "0"
	end
end

local function sendStatus(event)
	tcp:send(encode(event, positionNotNegative()))
end

local function eventPause(_, val)
	if val == true then
		sendStatus(EVENT_PAUSE)
	elseif val == false then
		sendStatus(EVENT_PLAY)
	end
end

local function eventSeek(_, val)
	if val == false then
		sendStatus(EVENT_SEEK)
	end
end

local function receiveReducer(event, pos)
	if event == EVENT_PLAY then
		mp.set_property_native("pause", false)
	elseif event == EVENT_PAUSE then
		mp.set_property_native("pause", true)
	elseif event == EVENT_SEEK then
		if math.abs(positionNotNegative() - pos) > SEEK_THRESHOLD then
			mp.set_property("time-pos", pos)
		end
	end
end


tcp:connect(host, port)

mp.observe_property("pause", "bool", eventPause)
mp.observe_property("seeking", "bool", eventSeek)

tcp:settimeout(0.05)

mp.add_periodic_timer(0.1, function()
	local msg = tcp:receive()
	if (msg ~= nil) then
		receiveReducer(decode(msg))
	end
end)
