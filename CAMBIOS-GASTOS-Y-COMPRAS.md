# EB Gestión — Gastos y corrección de compras sugeridas

Cambios incluidos:

1. Las compras sugeridas se generan únicamente cuando `stock < stock mínimo`.
2. La cantidad sugerida es `stock mínimo - stock actual`.
3. Nuevo módulo **Gastos** para Administrador/Bodega y Gerencia.
4. Las recepciones de compras crean automáticamente una cuenta por pagar.
5. La recepción permite registrar emisión, vencimiento, forma de pago, estado y fecha de pago.
6. Gastos manuales para servicios básicos y otros conceptos.
7. Alertas de gastos pendientes que vencen dentro de 7 días o están vencidos.

Archivos a reemplazar:
- `main.go`
- `web/index.html`
