# EB Gestión

Aplicación de inventario, compras, solicitudes, proveedores y control interno para Restaurante Entre Bahías.

## Objetivo de esta versión

- Ejecutarse en una ventana propia de Windows mediante Microsoft Edge WebView2.
- Mantener el servidor local activo para que teléfonos y otros equipos de la misma red puedan conectarse simultáneamente.
- Generar automáticamente `EBGestion.exe` y `EBGestion-Setup.exe` con GitHub Actions.

## Datos

Los datos no se guardan dentro de la carpeta de instalación. En Windows se almacenan en:

`%LOCALAPPDATA%\EBGestion\data.json`

Los respaldos se almacenan en:

`%LOCALAPPDATA%\EBGestion\respaldos`

Los archivos de datos están excluidos del repositorio para evitar publicar información operacional.

## Compilación automática

Cada subida a la rama `main` ejecuta el flujo **Construir EB Gestión para Windows**. Al finalizar, GitHub entrega un artefacto llamado `EB-Gestion-Windows` con:

- `EBGestion.exe`
- `EBGestion-Setup.exe`
- `LEEME-DESKTOP.txt`

También se puede ejecutar manualmente desde la pestaña **Actions**.
