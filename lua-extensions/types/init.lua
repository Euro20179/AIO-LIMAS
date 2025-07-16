---@meta

---@class Aio
---@field artStyles table<number, string>
---@field types string[]
---@field hasprefix fun(s: string, prefix: string): boolean [ Checks if a string has a prefix ]
---@field title fun(s: string): string [ Uppercases the first letter of the string ]
---@field listen fun(event: string, onevent: fun(...: any[]): any) [ Registers an event listener ]
---@field bor fun(n1: number, n2: number): number [ bitwise or 2 numbers ]
---@field split fun(s: string, sep: string): table<string>
---@field username2id fun(s: string): number
_G.aio = {}
