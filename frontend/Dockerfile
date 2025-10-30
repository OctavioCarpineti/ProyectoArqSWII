# Imagen base mínima
FROM nginx:alpine

# Crear un archivo HTML vacío para evitar errores de ruta
RUN mkdir -p /usr/share/nginx/html \
    && echo "<h1>Frontend temporal - React pendiente</h1>" > /usr/share/nginx/html/index.html

# Exponer el puerto (para mantener compatibilidad con docker-compose)
EXPOSE 80

# Comando por defecto
CMD ["nginx", "-g", "daemon off;"]