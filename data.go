package main

var pasteTemplate string = `
local files = {
%s
}

for _, file in ipairs(files) do 
    local fileName = file[1]
    local contents = file[2]
    local target = fs.open(shell.resolve(".").."/"..fileName, "w")

	io.write("Writing "..fileName.."...")

    local buffer = ""
    for _, byte in ipairs(contents) do 
        buffer = buffer .. string.char(byte)
    end
    target.write(buffer)
    target.close()

	io.write(" DONE")
end
print("Build finished")
`

var tomlTemplate string = `
DevApiKey = "YOUR DEV API KEY HERE"
SourceFolder = "src"
Username = "YOUR USERNAME HERE"
Password = "YOUR PASSWORD HERE"
`

var configName string = "config.plakbak"
