--wrk.method="GET"
--wrk.headers["Content-Type"]="application/json"
--wrk.headers["User-Agent"]="PostmanRuntime/7.32.3"
--wrk.headers["Authorization"]="Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTM5NjU1MjAsIlVpZCI6MSwiUmVmcmVzaENvdW50IjoxLCJVc2VyQWdlbnQiOiJQb3N0bWFuUnVudGltZS83LjMyLjMifQ.g98TFQ-ZYEzHromNYgZaSo_6NaO3cZT8VGRDeMIcW3B-Nz-4l0wjOMiR6NfQwV6S0H-_JwNwy6007Lj21KPxSA"

token = nil
path = "/user/login"
method = "POST"

wrk.headers["Content-Type"] = "application/json"
wrk.headers["User-Agent"] = ""

request = function ()
    body = '{"email": "tsukiyo6@163.com", "password": "for.nothing"}'
    return wrk.format(method, path, wrk.headers, body)
end

response = function (status, headers, body)
    if not token and status == 200 then
        token = headers["X-Jwt-Token"]
        path = "/user/profile"
        method = "GET"
        wrk.headers["Authorization"] = string.format("Bear %s", token)
    end
end