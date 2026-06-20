local opt = vim.opt

-- Line numbers
opt.number = true
opt.relativenumber = true

-- Tabs & indentation
opt.tabstop = 2
opt.shiftwidth = 2
opt.expandtab = true
opt.autoindent = true

-- Search
opt.ignorecase = true
opt.smartcase = true
opt.hlsearch = false
opt.incsearch = true

-- Appearance
opt.termguicolors = true
opt.signcolumn = "yes"
opt.cursorline = true
opt.scrolloff = 8
opt.wrap = false

-- Splits
opt.splitright = true
opt.splitbelow = true

-- Clipboard
opt.clipboard = "unnamedplus"

-- Undo
opt.undofile = true

-- Update time (for CursorHold, gitgutter)
opt.updatetime = 250

-- Ensure mise-managed runtimes are available to LSP servers
vim.env.PATH = vim.env.HOME .. "/.local/share/mise/shims:" .. vim.env.PATH
