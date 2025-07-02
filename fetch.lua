local http = require "resty.http"
local cjson = require "cjson"

-- Get username from URI
local username = ngx.var.username
if not username then
    ngx.status = 400
    ngx.say("Invalid request")
    return
end

-- Create HTTP client
local httpc = http.new()

-- Build target URL with query parameters
local target_url = "https://walletofsatoshi.com/.well-known/lnurlp/" .. username
if ngx.var.args then
    target_url = target_url .. "?" .. ngx.var.args
end

-- Make request to walletofsatoshi.com
local res, err = httpc:request_uri(target_url, {
    method = "GET",
    ssl_verify = true,
    timeout = 10000,
})

if not res then
    ngx.log(ngx.ERR, "Failed to request: ", err)
    ngx.status = 502
    ngx.say("Service temporarily unavailable")
    return
end

-- Set response headers
ngx.status = res.status
for k, v in pairs(res.headers) do
    if k ~= "connection" and k ~= "content-length" then
        ngx.header[k] = v
    end
end

-- Return the response body
ngx.say(res.body)