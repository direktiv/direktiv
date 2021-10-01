local plugin = {
  PRIORITY = 500,
  VERSION = "0.1",
}

function plugin:init_worker()
end


-- runs in the 'access_by_lua_block'
function plugin:access(plugin_conf)
  kong.log.inspect(plugin_conf)   -- check the logs for a pretty-printed config!
end


-- runs in the 'header_filter_by_lua_block'
function plugin:header_filter(plugin_conf)
  kong.response.set_header("Access-Control-Allow-Origin", "*")
  kong.response.set_header("Access-Control-Allow-Headers", "Content-Type")
  kong.response.set_header("Content-Type",  "text/event-stream")
  kong.response.set_header("Cache-Control", "no-cache")
  kong.response.set_header("Connection", "keep-alive")

end --]]

local concat = table.concat
local upper = string.upper

function plugin:body_filter(conf)
  local chunk, eof = ngx.arg[1], ngx.arg[2]
  local ctx = ngx.ctx

  ctx.rt_body_chunks = ctx.rt_body_chunks or {}
  ctx.rt_body_chunk_number = ctx.rt_body_chunk_number or 1

  -- kong.log.debug("!!!! GOT CHUNK === ", chunk or "NO Chunk!")

  -- Check if response is bonked 
  -- kong.log.debug("!!!! response === ", kong.response.get_status())
  -- kong.log.debug("!!!! CONTENT TYPE === ", kong.response.get_header("Content-Type"))
  -- kong.log.debug("!!!! grpc-message === ", kong.response.get_header("grpc-message") or "Internal Error")

  if kong.response.get_status() ~= 200 then
    kong.log.debug("Got error from server grpc")
    local errorMsg = kong.response.get_header("grpc-message") or "Internal Error"
    chunk = string.format("error: %s\n\n", errorMsg)
    kong.log.debug("Attempting to set proccessed error: ", chunk)
    ctx.rt_body_chunks[ctx.rt_body_chunk_number] = chunk
    ctx.rt_body_chunk_number = ctx.rt_body_chunk_number + 1

    ngx.arg[1] = chunk
  elseif eof then
    kong.log.debug("Reached end of file with chunk: ", chunk)

    local body = concat(ctx.rt_body_chunks)
    -- kong.log.debug("body === ", body)
    -- ngx.arg[1] = upper(body)
    ngx.arg[1] = "error: eof\n\n"
  else
    ctx.rt_body_chunks[ctx.rt_body_chunk_number] = chunk
    ctx.rt_body_chunk_number = ctx.rt_body_chunk_number + 1

    -- kong.log.debug("!!!! GOT ctx.rt_body_chunks === ", ctx.rt_body_chunks or "NO ctx.rt_body_chunks!")
    -- kong.log.debug("!!!! GOT ctx.rt_body_chunk_number === ", ctx.rt_body_chunk_number or "NO ctx.rt_body_chunk_number!")

    ngx.arg[1] = string.format("data: %s\n\n", chunk)
  end
end

return plugin
