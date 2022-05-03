module.exports = {
  title: 'CassKop',
  tagline: 'Open-Source, Apache Cassandra operator for Kubernetes',
  url: 'https://cscetbon.github.io',
  baseUrl: '/casskop/',
  onBrokenLinks: 'throw',
  favicon: 'img/casskop.ico',
  organizationName: 'cscetbon',
  projectName: 'casskop',
  themeConfig: {
    navbar: {
      title: 'CassKop',
      logo: {
        alt: 'CassKop Logo',
        src: 'img/casskop_alone.png',
      },
      items: [
        {to: 'docs/concepts/introduction', label: 'Docs', position: 'right'},
        {to: 'blog', label: 'Blog', position: 'right'},
        {
          href: 'https://github.com/cscetbon/casskop',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Getting Started',
          items: [
            {
              label: 'Documentation',
              to: 'docs/concepts/introduction',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/cscetbon/casskop',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Slack',
              href: 'https://casskop.slack.com',
            },
            {
              label: 'Blog',
              to: 'blog',
            },
          ],
        },
        {
          title: 'Contact',
          items: [
            {
              label: 'Feature request',
              href: 'https://github.com/cscetbon/casskop/issues',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Orange, Inc. Built with Docusaurus.`,
    },

    // Search option
    algolia: {
      appId: 'N8HRRYXE70',
      apiKey: '20115d6011851f46b33f7ffc1960153c',
      indexName: 'casskop',
    },
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl:
            'https://github.com/cscetbon/casskop/edit/master/website/',
        },
        blog: {
          showReadingTime: true,
          editUrl:
            'https://github.com/cscetbon/casskop/edit/master/website/blog',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
