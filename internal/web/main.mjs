/**
* @typedef {Object} OSData
* @property {string} id
* @property {string[]} [id_like]
* @property {string} os
* @property {string} arch
* @property {string} kernel
* @property {string} uptime
* @property {string} hostname
* @property {string} device
*
* @typedef {Object} CPUData
* @property {string} brand
* @property {string} model
* @property {number} cores
* @property {number} threads
* @property {number} frequency
*
* @typedef {Object} MemData
* @property {number} usedMiB
* @property {number} totalMiB
* @property {number} usedPercent
*
* @typedef {Object} StorageData
* @property {string} device
* @property {string} mountpoint
* @property {string} fsType
* @property {string} type
* @property {number} totalMiB
* @property {number} usedMiB
* @property {number} freeMiB
* @property {number} usedPercent
*
* @typedef {Object} SensorReading
* @property {string} label
* @property {number} value
* @property {string} [unit]
* @property {string} [extra]
*
* @typedef {Object} SensorChip
* @property {string} name
* @property {string} adapter
* @property {SensorReading[]} readings
*/

class DoveDashUI {
	/**
	* Removes leading and trailing whitespace and empty lines
	* @param {string} str
	* @returns {string}
	*/
	static dedent(str) {
		return str.replace(/^\s*\n/, '').replace(/\n\s*$/, '').replace(/^[ \t]+/gm, '')
	}

	/**
	* Normalizes a distro identifier for icon lookups
	* @param {string | undefined} value
	* @returns {string}
	*/
	static normalizeIconId(value) {
		return (value || '')
			.toLowerCase()
			.trim()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/^-+|-+$/g, '')
	}

	/**
	* Selects a distro-specific icon path
	* @param {OSData} data
	* @returns {string}
	*/
	static getDistroIconPath(data) {
		const aliases = {
			ubuntu: 'ubuntu',
			kubuntu: 'ubuntu',
			xubuntu: 'ubuntu',
			lubuntu: 'ubuntu',
			debian: 'debian',
			raspbian: 'debian',
			fedora: 'fedora',
			rhel: 'fedora',
			redhat: 'fedora',
			centos: 'fedora',
			rocky: 'fedora',
			almalinux: 'fedora',
			arch: 'archlinux',
			archlinux: 'archlinux',
			opensuse: 'opensuse',
			suse: 'opensuse',
			linuxmint: 'linuxmint',
			mint: 'linuxmint',
			manjaro: 'manjaro',
			alpine: 'alpine',
			kali: 'kali',
			pop: 'popos',
			popos: 'popos',
			nix: 'nixos',
			nixos: 'nixos',
		}

		const searchTerms = [data.id, data.os, ...(data.id_like || [])]
		for (const term of searchTerms) {
			const normalized = this.normalizeIconId(term)
			if (!normalized) continue
			const alias = aliases[normalized]
			if (alias) {
				return `https://cdn.jsdelivr.net/npm/simple-icons@11.15.0/icons/${alias}.svg`
			}
		}

		return 'https://cdn.jsdelivr.net/npm/simple-icons@11.15.0/icons/linux.svg'
	}

	/**
	* Loads the distro icon as inline SVG from a remote URL
	* @param {OSData} data
	* @returns {Promise<string>}
	*/
	static async loadDistroIcon(data) {
		const iconPath = DoveDashUI.getDistroIconPath(data)
		try {
			const res = await fetch(iconPath, { headers: { Accept: 'image/svg+xml' }, cache: 'no-store' })
			if (!res.ok) {
				return ''
			}
			const svg = await res.text()

			return svg
				.replace(/^<\?xml[\s\S]*?\?>\s*/i, '')
				.replace(/^<!doctype[\s\S]*?>\s*/i, '')
				.replace(/fill="#([^"]*)"/gi, 'fill="currentColor"')
				.replace(/fill='#[^']*'/gi, 'fill="currentColor"')
				.replace(/stroke="#([^"]*)"/gi, 'stroke="currentColor"')
				.replace(/stroke='[^']*'/gi, 'stroke="currentColor"')
				.replace(/<svg\b([^>]*)>/i, '<svg class="distro-icon" focusable="false" aria-hidden="true"$1>')
				.replace(/\swidth="[^"]*"/gi, '')
				.replace(/\sheight="[^"]*"/gi, '')
		} catch (err) {
			console.error('Failed to load distro icon', err)
			return ''
		}
	}

	/**
	* Mebibytes to Gibibytes conversion
	* @param {number} mib
	* @returns {string}
	*/
	static toGiB(mib) {
		return (mib / 1024).toFixed(2)
	}

	/**
	* Formats a time difference in seconds into a human-readable string
	* @param {number} secondsAgo
	* @returns {string}
	*/
	static formatTimeAgo(secondsAgo) {
		if (secondsAgo <= 60) return 'less than a minute ago'
		const minutes = Math.floor(secondsAgo / 60)
		return `${minutes} minute${minutes !== 1 ? 's' : ''} ago`
	}

	/**
	* Downloads data from the given endpoint and formats it using the provided formatter function
	* @template T
	* @param {string} endpoint
	* @param {string} elementId
	* @param {(data: T) => string | Promise<string>} formatter
	* @returns {Promise<void>}
	*/
	static async fetchAndDisplay(endpoint, elementId, formatter) {
		try {
			const el = document.getElementById(elementId)
			if (!el.dataset.loaded) {
				el.innerHTML = `<p class="info-line">Loading data...</p>`
			}

			const res = await fetch(endpoint)
			const data = await res.json()
			const formatted = await formatter(data)

			el.innerHTML = formatted
			el.dataset.loaded = true

			const section = el.closest('section')
			if (section) {
				section.classList.remove('loading')
				section.classList.add('loaded')
			}
		} catch (err) {
			const el = document.getElementById(elementId)
			el.innerHTML = `<p class="error-line">Error loading data</p>`
			console.error(err)
		}
	}

	/**
	* Formats the OS data into HTML
	* @param {OSData} data
	* @returns {Promise<string>}
	*/
	static async formatOS(data) {
		const iconSvg = await DoveDashUI.loadDistroIcon(data)

		return DoveDashUI.dedent(`
			<div class="info-block">
				<div class="info-header">
					<div class="info-icon">${iconSvg}</div>
					<span class="info-name">${data.os}</span>
				</div>
				<div class="info-list">
					<p class="info-line"><strong>Architecture:</strong> ${data.arch}</p>
					<p class="info-line"><strong>Kernel:</strong> ${data.kernel}</p>
					<p class="info-line"><strong>Uptime:</strong> ${data.uptime}</p>
					<p class="info-line"><strong>Hostname:</strong> ${data.hostname}</p>
					<p class="info-line"><strong>Device:</strong> ${data.device}</p>
				</div>
			</div>
		`)
	}

	/**
	* Formats the CPU data into HTML
	* @param {CPUData} data
	* @returns {string}
	*/
	static formatCPU(data) {
		return DoveDashUI.dedent(`
			<div class="info-list">
				<p class="info-line"><strong>Name:</strong> ${data.name}</p>
				<p class="info-line"><strong>Cores/Threads:</strong> ${data.cores}/${data.threads}</p>
				<p class="info-line"><strong>Frequency:</strong> ${data.frequency} GHz</p>
			</div>
		`)
	}

	/**
	* Formats the memory data into HTML
	* @param {MemData} data
	* @returns {string}
	*/
	static formatMem(data) {
		const used = isFinite(data.usedMiB) ? DoveDashUI.toGiB(data.usedMiB) : '?'
		const total = isFinite(data.totalMiB) ? DoveDashUI.toGiB(data.totalMiB) : '?'
		const percent = isFinite(data.usedPercent) ? data.usedPercent.toFixed(0) : '?'

		return `<p class="info-line"><strong>Usage:</strong> ${used} GiB / ${total} GiB (${percent}%)</p>`
	}

	/**
	* Formats the storage data into HTML 
	* @param {StorageData} data
	* @returns {string}
	*/
	static formatStorage(data) {
		const used = isFinite(data.usedMiB) ? DoveDashUI.toGiB(data.usedMiB) : '?'
		const total = isFinite(data.totalMiB) ? DoveDashUI.toGiB(data.totalMiB) : '?'
		const percent = isFinite(data.usedPercent) ? data.usedPercent.toFixed(2) : '?'
		const mount = data.mountpoint || '/'
		const fs = data.fsType || 'unknown'

		return DoveDashUI.dedent(`
			<div class="info-list">
				<p class="info-line"><strong>Type and filesystem:</strong> ${data.type}, ${fs}</p>
				<p class="info-line"><strong>Disk (${mount}):</strong> ${used} GiB / ${total} GiB (${percent}%)</p>
			</div>
		`)
	}

	/**
	* Formats the sensor data into HTML
	* @param {SensorChip[]} data
	* @returns {string}
	*/
	static formatSensors(data) {
		return DoveDashUI.dedent(`
			<div class="sensors-list">
				${data.map(chip => `
					<div class="info-block">
						<div class="info-list">
							<h3 class="info-name">${chip.name}</h3>
							<p class="info-line"><strong>Adapter:</strong> ${chip.adapter}</p>
							${chip.readings.map(r => {
								let tempClass = ""
								if (r.unit === "°C") {
									if (r.value < 30) tempClass = "temp-info"
									else if (r.value < 60) tempClass = "temp-success"
									else if (r.value < 80) tempClass = "temp-warning"
									else tempClass = "temp-error"
								}
								return `<p class="info-line ${tempClass}"><strong>${r.label}:</strong> ${r.value.toFixed(1)}${r.unit || ''} ${r.extra ? `<span class="sensor-extra">${r.extra}</span>` : ''}</p>`
							}).join('')}
						</div>
					</div>
				`).join('')}
			</div>
		`)
	}

	/**
	* Formats the network data into HTML
	* @param {NetStats[]} data
	* @returns {string}
	*/
	static formatNet(data) {
		const net = data[0]

		const down = net.speedDownMbps.toFixed(2)
		const up = net.speedUpMbps.toFixed(2)
		const timeAgo = DoveDashUI.formatTimeAgo(Math.floor((Date.now() - new Date(net.lastBenchmark)) / 1000))
		const interfaceBandwidth = net.bandwidth && net.bandwidth > 0
			? (net.bandwidth >= 1000
				? `${(net.bandwidth / 1000).toFixed(1)} Gb/s`
				: `${net.bandwidth.toFixed(1)} Mb/s`)
			: "No info"

		return DoveDashUI.dedent(`
			<div class="info-list">
				<p class="info-line"><strong>Interface:</strong> ${net.name} (${net.type})</p>
				<p class="info-line"><strong>Bandwidth:</strong> ${interfaceBandwidth} </p>
				<p class="info-line"><strong>Download/Upload:</strong> ↓ ${down} Mb/s / ↑ ${up} Mb/s </p>
				<p class="info-line"><strong>Last benchmark:</strong> ${new Date(net.lastBenchmark).toLocaleString('en-GB')} (${timeAgo})</p>
			</div>
		`)
	}

	/**
	* Refreshes all data by fetching from the API and displaying it
	* @returns {Promise<void>}
	*/
	static refreshAll() {
		DoveDashUI.fetchAndDisplay('/api/os', 'os-data', DoveDashUI.formatOS)
		DoveDashUI.fetchAndDisplay('/api/cpu', 'cpu-data', DoveDashUI.formatCPU)
		DoveDashUI.fetchAndDisplay('/api/mem', 'mem-data', DoveDashUI.formatMem)
		DoveDashUI.fetchAndDisplay('/api/storage', 'storage-data', DoveDashUI.formatStorage)
		DoveDashUI.fetchAndDisplay('/api/sensors', 'sensors-data', DoveDashUI.formatSensors)
		// DoveDashUI.fetchAndDisplay('/api/net', 'net-data', DoveDashUI.formatNet)
	}
}

DoveDashUI.refreshAll()
setInterval(() => DoveDashUI.refreshAll(), 10000)
