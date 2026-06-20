require("mason").setup()
require("mason-lspconfig").setup({
  ensure_installed = { "lua_ls", "pyright", "gopls", "rust_analyzer", "ts_ls" },
  automatic_installation = true,
})

local capabilities = vim.tbl_deep_extend(
  "force",
  vim.lsp.protocol.make_client_capabilities(),
  require("cmp_nvim_lsp").default_capabilities()
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

require("mason-lspconfig").setup_handlers({
  function(server_name)
    require("lspconfig")[server_name].setup({
      capabilities = capabilities,
      on_attach = on_attach,
    })
  end,
})
