import { VitePWA } from 'vite-plugin-pwa'

export default VitePWA({
	devOptions: {
		enabled: true,
		type: 'module',
	},
	srcDir: 'src',
	filename: 'sw.ts',
	registerType: 'autoUpdate',
	strategies: 'injectManifest', // for custom SW
	includeAssets: ['favicon.ico'],
	manifest: {
		name: 'GoOpenBmclApi Dashboard',
		short_name: 'GOBA Dash',
		description: 'Go-Openbmclapi Internal Dashboard',
		theme_color: '#4c89fe',
		icons: [
			{
				src: 'pwa-64x64.png',
				sizes: '64x64',
				type: 'image/png',
			},
			{
				src: 'pwa-192x192.png',
				sizes: '192x192',
				type: 'image/png',
			},
			{
				src: 'pwa-512x512.png',
				sizes: '512x512',
				type: 'image/png',
				purpose: 'any',
			},
			{
				src: 'maskable-icon-512x512.png',
				sizes: '512x512',
				type: 'image/png',
				purpose: 'maskable',
			},
			{
				src: 'logo.png',
				sizes: '512x512',
				type: 'image/png',
			},
		],
	},
	workbox: {
		globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
	},
})
