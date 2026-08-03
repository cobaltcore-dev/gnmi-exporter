import { withMermaid } from 'vitepress-plugin-mermaid'
import { fileURLToPath, URL } from 'node:url'

// https://vitepress.dev/reference/site-config
export default withMermaid({
    title: 'gNMI Exporter',
    description: 'Cloud Native Network Device Monitoring',
    base: '/gnmi-exporter/',
    head: [
        [
            'link',
            {
                rel: 'icon',
                href: 'https://raw.githubusercontent.com/ironcore-dev/ironcore/refs/heads/main/docs/assets/logo_borderless.svg',
            },
        ],
    ],
    vite: {
        resolve: {
            alias: [
                {
                    find: /^.*\/VPFooter\.vue$/,
                    replacement: fileURLToPath(new URL('./theme/components/VPFooter.vue', import.meta.url)),
                },
            ],
        },
    },
    themeConfig: {
        // https://vitepress.dev/reference/default-theme-config
        nav: [
            { text: 'Home', link: '/' },
            {
                text: 'Documentation',
                items: [
                    { text: 'Overview', link: '/overview/' },
                    { text: 'API Reference', link: '/api-reference/' },
                ],
            },
            {
                text: 'Projects',
                items: [
                    { text: 'ApeiroRA', link: 'https://apeirora.eu/' },
                    { text: 'IronCore', link: 'https://ironcore.dev/' },
                    {
                        text: 'CobaltCore',
                        link: 'https://cobaltcore-dev.github.io/docs/',
                    },
                ],
            },
        ],

        editLink: {
            pattern: 'https://github.com/cobaltcore-dev/gnmi-exporter/blob/main/docs/:path',
            text: 'Edit this page on GitHub',
        },

          logo: {
            src: 'https://raw.githubusercontent.com/ironcore-dev/ironcore/refs/heads/main/docs/assets/logo_borderless.svg',
            width: 24,
            height: 24,
        },

        search: {
            provider: 'local',
        },

        sidebar: [
            {
                text: 'Overview',
                items: [
                    { text: 'Introduction', link: '/overview/' },
                    { text: 'Quickstart', link: '/overview/quickstart' },
                    { text: 'Architecture', link: '/overview/architecture' },
                ],
            },
            {
                text: 'API Reference',
                items: [{ text: 'DeviceMonitor', link: '/api-reference/' }],
            },
        ],

        socialLinks: [{ icon: 'github', link: 'https://github.com/cobaltcore-dev/gnmi-exporter' }],
    },
})
