# EB Gestión Build 0.9.1

## Cambio incluido

- Compras sugeridas: solo muestra productos cuyo stock actual es menor al stock mínimo.
- Cantidad sugerida: stock mínimo menos stock actual.
- Los productos con stock igual al mínimo ya no aparecen.

## Ejemplos

- Stock 1 / mínimo 1: no sugerir.
- Stock 0 / mínimo 1: sugerir 1.
- Stock 1 / mínimo 6: sugerir 5.
- Stock 20 / mínimo 20: no sugerir.

Esta compilación no incluye todavía el módulo Gastos. Se separó para mantener el cambio pequeño y seguro.
