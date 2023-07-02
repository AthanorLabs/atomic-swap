
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * Environment variables [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env`. Like [`$env/dynamic/private`](https://kit.svelte.dev/docs/modules#$env-dynamic-private), this module cannot be imported into client-side code. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://kit.svelte.dev/docs/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://kit.svelte.dev/docs/configuration#env) (if configured).
 * 
 * _Unlike_ [`$env/dynamic/private`](https://kit.svelte.dev/docs/modules#$env-dynamic-private), the values exported from this module are statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * ```ts
 * import { API_KEY } from '$env/static/private';
 * ```
 * 
 * Note that all environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * 
 * ```
 * MY_FEATURE_FLAG=""
 * ```
 * 
 * You can override `.env` values from the command line like so:
 * 
 * ```bash
 * MY_FEATURE_FLAG="enabled" npm run dev
 * ```
 */
declare module '$env/static/private' {
	export const SHELL: string;
	export const LSCOLORS: string;
	export const WINDOWID: string;
	export const COLORTERM: string;
	export const LESS: string;
	export const NVM_INC: string;
	export const CONDA_EXE: string;
	export const _CE_M: string;
	export const NVIM: string;
	export const npm_package_dependencies_ethers: string;
	export const npm_package_dependencies_flowbite: string;
	export const I3SOCK: string;
	export const NODE: string;
	export const npm_config_ignore_scripts: string;
	export const npm_package_devDependencies_autoprefixer: string;
	export const npm_package_devDependencies_tailwindcss: string;
	export const npm_package_dependencies__popperjs_core: string;
	export const npm_package_private: string;
	export const npm_package_devDependencies_svelte_heros_v2: string;
	export const npm_config_argv: string;
	export const NVIM_LOG_FILE: string;
	export const npm_config_bin_links: string;
	export const KITTY_PID: string;
	export const MASON: string;
	export const npm_package_dependencies_classnames: string;
	export const GTK_MODULES: string;
	export const XDG_SEAT: string;
	export const npm_package_dependencies_axios: string;
	export const PWD: string;
	export const npm_config_save_prefix: string;
	export const npm_package_devDependencies_vite: string;
	export const LOGNAME: string;
	export const XDG_SESSION_TYPE: string;
	export const npm_package_readmeFilename: string;
	export const npm_package_dependencies_svelte_awesome_icons: string;
	export const npm_package_scripts_build: string;
	export const _: string;
	export const XAUTHORITY: string;
	export const DESKTOP_STARTUP_ID: string;
	export const KITTY_PUBLIC_KEY: string;
	export const MOTD_SHOWN: string;
	export const HOME: string;
	export const npm_config_version_git_tag: string;
	export const LANG: string;
	export const LS_COLORS: string;
	export const npm_config_init_license: string;
	export const npm_package_version: string;
	export const VIMRUNTIME: string;
	export const npm_config_version_commit_hooks: string;
	export const KITTY_WINDOW_ID: string;
	export const npm_package_dependencies__sveltejs_kit: string;
	export const INIT_CWD: string;
	export const npm_package_scripts_preview: string;
	export const npm_lifecycle_script: string;
	export const npm_package_description: string;
	export const NVM_DIR: string;
	export const npm_config_version_tag_prefix: string;
	export const npm_package_devDependencies__sveltejs_vite_plugin_svelte: string;
	export const YARN_WRAP_OUTPUT: string;
	export const XDG_SESSION_CLASS: string;
	export const TERM: string;
	export const TERMINFO: string;
	export const npm_package_name: string;
	export const ZSH: string;
	export const _CE_CONDA: string;
	export const npm_package_type: string;
	export const USER: string;
	export const npm_package_dependencies_flowbite_svelte: string;
	export const CONDA_SHLVL: string;
	export const DISPLAY: string;
	export const npm_lifecycle_event: string;
	export const SHLVL: string;
	export const npm_config_version_git_sign: string;
	export const NVM_CD_FLAGS: string;
	export const npm_config_version_git_message: string;
	export const PAGER: string;
	export const XDG_VTNR: string;
	export const XDG_SESSION_ID: string;
	export const npm_config_user_agent: string;
	export const npm_execpath: string;
	export const CONDA_PYTHON_EXE: string;
	export const npm_package_devDependencies_svelte: string;
	export const XDG_RUNTIME_DIR: string;
	export const npm_package_devDependencies_svelte_preprocess: string;
	export const MYVIMRC: string;
	export const npm_config_strict_ssl: string;
	export const DEBUGINFOD_URLS: string;
	export const npm_package_scripts_dev: string;
	export const PATH: string;
	export const npm_config_node_gyp: string;
	export const npm_package_dependencies_identicon_js: string;
	export const DBUS_SESSION_BUS_ADDRESS: string;
	export const NVM_BIN: string;
	export const MAIL: string;
	export const npm_config_registry: string;
	export const npm_config_ignore_optional: string;
	export const npm_config_scripts_prepend_node_path: string;
	export const npm_package_devDependencies_postcss: string;
	export const KITTY_INSTALLATION_DIR: string;
	export const npm_node_execpath: string;
	export const npm_package_devDependencies_postcss_load_config: string;
	export const OLDPWD: string;
	export const npm_config_init_version: string;
	export const NODE_ENV: string;
}

/**
 * Similar to [`$env/static/private`](https://kit.svelte.dev/docs/modules#$env-static-private), except that it only includes environment variables that begin with [`config.kit.env.publicPrefix`](https://kit.svelte.dev/docs/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Values are replaced statically at build time.
 * 
 * ```ts
 * import { PUBLIC_BASE_URL } from '$env/static/public';
 * ```
 */
declare module '$env/static/public' {
	
}

/**
 * This module provides access to runtime environment variables, as defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/master/packages/adapter-node) (or running [`vite preview`](https://kit.svelte.dev/docs/cli)), this is equivalent to `process.env`. This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://kit.svelte.dev/docs/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://kit.svelte.dev/docs/configuration#env) (if configured).
 * 
 * This module cannot be imported into client-side code.
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * console.log(env.DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 * 
 * > In `dev`, `$env/dynamic` always includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 */
declare module '$env/dynamic/private' {
	export const env: {
		SHELL: string;
		LSCOLORS: string;
		WINDOWID: string;
		COLORTERM: string;
		LESS: string;
		NVM_INC: string;
		CONDA_EXE: string;
		_CE_M: string;
		NVIM: string;
		npm_package_dependencies_ethers: string;
		npm_package_dependencies_flowbite: string;
		I3SOCK: string;
		NODE: string;
		npm_config_ignore_scripts: string;
		npm_package_devDependencies_autoprefixer: string;
		npm_package_devDependencies_tailwindcss: string;
		npm_package_dependencies__popperjs_core: string;
		npm_package_private: string;
		npm_package_devDependencies_svelte_heros_v2: string;
		npm_config_argv: string;
		NVIM_LOG_FILE: string;
		npm_config_bin_links: string;
		KITTY_PID: string;
		MASON: string;
		npm_package_dependencies_classnames: string;
		GTK_MODULES: string;
		XDG_SEAT: string;
		npm_package_dependencies_axios: string;
		PWD: string;
		npm_config_save_prefix: string;
		npm_package_devDependencies_vite: string;
		LOGNAME: string;
		XDG_SESSION_TYPE: string;
		npm_package_readmeFilename: string;
		npm_package_dependencies_svelte_awesome_icons: string;
		npm_package_scripts_build: string;
		_: string;
		XAUTHORITY: string;
		DESKTOP_STARTUP_ID: string;
		KITTY_PUBLIC_KEY: string;
		MOTD_SHOWN: string;
		HOME: string;
		npm_config_version_git_tag: string;
		LANG: string;
		LS_COLORS: string;
		npm_config_init_license: string;
		npm_package_version: string;
		VIMRUNTIME: string;
		npm_config_version_commit_hooks: string;
		KITTY_WINDOW_ID: string;
		npm_package_dependencies__sveltejs_kit: string;
		INIT_CWD: string;
		npm_package_scripts_preview: string;
		npm_lifecycle_script: string;
		npm_package_description: string;
		NVM_DIR: string;
		npm_config_version_tag_prefix: string;
		npm_package_devDependencies__sveltejs_vite_plugin_svelte: string;
		YARN_WRAP_OUTPUT: string;
		XDG_SESSION_CLASS: string;
		TERM: string;
		TERMINFO: string;
		npm_package_name: string;
		ZSH: string;
		_CE_CONDA: string;
		npm_package_type: string;
		USER: string;
		npm_package_dependencies_flowbite_svelte: string;
		CONDA_SHLVL: string;
		DISPLAY: string;
		npm_lifecycle_event: string;
		SHLVL: string;
		npm_config_version_git_sign: string;
		NVM_CD_FLAGS: string;
		npm_config_version_git_message: string;
		PAGER: string;
		XDG_VTNR: string;
		XDG_SESSION_ID: string;
		npm_config_user_agent: string;
		npm_execpath: string;
		CONDA_PYTHON_EXE: string;
		npm_package_devDependencies_svelte: string;
		XDG_RUNTIME_DIR: string;
		npm_package_devDependencies_svelte_preprocess: string;
		MYVIMRC: string;
		npm_config_strict_ssl: string;
		DEBUGINFOD_URLS: string;
		npm_package_scripts_dev: string;
		PATH: string;
		npm_config_node_gyp: string;
		npm_package_dependencies_identicon_js: string;
		DBUS_SESSION_BUS_ADDRESS: string;
		NVM_BIN: string;
		MAIL: string;
		npm_config_registry: string;
		npm_config_ignore_optional: string;
		npm_config_scripts_prepend_node_path: string;
		npm_package_devDependencies_postcss: string;
		KITTY_INSTALLATION_DIR: string;
		npm_node_execpath: string;
		npm_package_devDependencies_postcss_load_config: string;
		OLDPWD: string;
		npm_config_init_version: string;
		NODE_ENV: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * Similar to [`$env/dynamic/private`](https://kit.svelte.dev/docs/modules#$env-dynamic-private), but only includes variables that begin with [`config.kit.env.publicPrefix`](https://kit.svelte.dev/docs/configuration#env) (which defaults to `PUBLIC_`), and can therefore safely be exposed to client-side code.
 * 
 * Note that public dynamic environment variables must all be sent from the server to the client, causing larger network requests — when possible, use `$env/static/public` instead.
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.PUBLIC_DEPLOYMENT_SPECIFIC_VARIABLE);
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
