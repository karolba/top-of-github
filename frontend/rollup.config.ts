import commonjs from '@rollup/plugin-commonjs';
import nodeResolve from '@rollup/plugin-node-resolve';
import typescript from '@rollup/plugin-typescript';
import serve from 'rollup-plugin-serve';
import html from '@rollup/plugin-html';
import del from 'rollup-plugin-delete';
import copy from 'rollup-plugin-copy';
import { InputPluginOption, RollupOptions } from 'rollup';
import { readFileSync } from 'fs';

const commonConfig: RollupOptions = {
    input: 'src/index.ts',
    output: {
        dir: 'dist',
        format: 'iife',
        sourcemap: true,
        validate: true,
        compact: true,
        entryFileNames: '[name]-[hash].js',
    },
};

const commonPlugins: InputPluginOption[] = [
    del({
        targets: 'dist/',
        runOnce: true,
        verbose: true,
    }),
    nodeResolve(),
    commonjs(),
    html({
        publicPath: '/',
        template(options): string {
            const scripts = (options?.files.js || []).map(({ fileName }) => `<script src="${options?.publicPath}${fileName}"></script>`)
            const links = (options?.files.css || []).map(({ fileName }) => `<link href="${options?.publicPath}${fileName}" rel="stylesheet">`)

            return readFileSync('src/index-template.html')
                .toString()
                .replace('${links}', links.join('\n'))
                .replace('${scripts}', scripts.join('\n'))
        },
    }),
    copy({
        targets: [
            { src: 'static/*', dest: 'dist/' }
        ]
    })
]

const configurations: Record<string, () => Promise<RollupOptions>> = {
    async DEV(): Promise<RollupOptions> {
        // Livereload cannot be imported on the top of the file, because it spawns
        // an external process on import and that makes the "PRODUCTION" config unable
        // to ever finish, hanging forever.
        // See https://github.com/rollup/rollup/issues/3827
        const livereload = await import('rollup-plugin-livereload');

        return {
            ...commonConfig,
            watch: {
                include: ['static/**', 'src/**', '*.ts'],
                exclude: 'dist',
            },
            plugins: [
                typescript(),
                ...commonPlugins,
                serve({
                    contentBase: 'dist',
                }),
                livereload.default(),
            ],
        };
    },
    async PRODUCTION(): Promise<RollupOptions> {
        return {
            ...commonConfig,
            plugins: [
                typescript({ noEmitOnError: true }),
                ...commonPlugins,
            ],
        };
    },
};

const selectedConfig = configurations[process.env.BUILD!];

if (!selectedConfig) {
    throw new Error(`The BUILD environment variable needs to be set to either of ${Object.keys(configurations)}`);
}

export default await selectedConfig();
