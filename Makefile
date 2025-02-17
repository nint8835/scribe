css:
	tailwindcss -i pkg/web/static/input.css -o pkg/web/static/tailwind.css

prod-css:
	tailwindcss -i pkg/web/static/input.css -o pkg/web/static/tailwind.css --minify
