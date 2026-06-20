local ok_mason, mason = pcall(require, "mason")
if not ok_mason then return end
mason.setup()

local ok_mlsp, mason_lspconfig = pcall(require, "mason-lspconfig")
if not ok_mlsp then return end
mason_lspconfig.setup({
  ensure_installed = { "lua_ls", "pyright", "gopls", "rust_analyzer", "ts_ls" },
  automatic_installation = true,
})

local ok_cmp, cmp_nvim_lsp = pcall(require, "cmp_nvim_lsp")
if not ok_cmp then return end

local capabilities = vim.tbl_deep_extend(
  "force",
  vim.lsp.protocol.make_client_capabilities(),
  cmp_nvim_lsp.default_capabilities()
)

local on_attach = function(_, bufnr)
  local map = function(keys, func)
    vim.keymap.set("n", keys, func, { buffer = bufnr })
  end
  map("gd", vim.lsp.buf.definition)
  map("gr", vim.lsp.buf.references)
  map("K", vim.lsp.buf.hover)
  map("<leader>rn", vim.lsp.buf.rename)
  map("<leader>ca", vim.lsp.buf.code_action)
  map("<leader>f", function() vim.lsp.buf.format({ async = true }) end)
end

-- Use new vim.lsp.config API (nvim 0.11+)
vim.lsp.config("*", { capabilities = capabilities, on_attach = on_attach })
vim.lsp.enable({ "lua_ls", "pyright", "gopls", "rust_analyzer", "ts_ls" })
