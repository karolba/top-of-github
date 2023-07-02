import commonjs from '@rollup/plugin-commonjs';
import nodeResolve from '@rollup/plugin-node-resolve';
import typescript from '@rollup/plugin-typescript';
import serve from 'rollup-plugin-serve';
import { RollupOptions } from 'rollup';

const commonConfig: RollupOptions = {
    input: 'src/index.ts',
    output: {
        file: 'static/dist/bundle.js',
        format: 'iife',
        sourcemap: true,
        validate: true,
        compact: true
    },
};

const configurations: Record<string, () => Promise<RollupOptions>> = {
    DEV: async function (): Promise<RollupOptions> {
        // Livereload cannot be imported on the top of the file, because it spawns
        // an external process on import and that makes the "PRODUCTION" config unable
        // to ever finish, hanging forever.
        // See https://github.com/rollup/rollup/issues/3827
        const livereload = await import('rollup-plugin-livereload');

        return {
            ...commonConfig,
            watch: {
                exclude: 'static/dist'
            },
            plugins: [
                typescript(),
                nodeResolve(),
                commonjs(),
                serve({
                    contentBase: 'static',
                }),
                livereload.default(),
            ],
        };
    },
    PRODUCTION: async function (): Promise<RollupOptions> {
        return {
            ...commonConfig,
            plugins: [
                typescript({ noEmitOnError: true }),
                nodeResolve(),
                commonjs(),
            ],
        };
    },
};

const selectedConfig = configurations[process.env.BUILD!];

if (!selectedConfig) {
    throw new Error(`The BUILD environment variable needs to be set to either of ${Object.keys(configurations)}`);
}

export default await selectedConfig();
