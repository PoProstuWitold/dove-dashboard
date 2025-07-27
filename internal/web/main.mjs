/**
* @typedef {Object} OSData
* @property {string} id
* @property {string} os
* @property {string} arch
* @property {string} kernel
* @property {string} uptime
* @property {string} hostname
*
* @typedef {Object} CPUData
* @property {string} brand
* @property {string} model
* @property {number} cores
* @property {number} threads
* @property {number} frequency
*
* @typedef {Object} MemData
* @property {number} usedMB
* @property {number} totalMB
* @property {number} usedPercent
*
* @typedef {Object} StorageData
* @property {string} device
* @property {string} mountpoint
* @property {string} fsType
* @property {string} type
* @property {number} totalMB
* @property {number} usedMB
* @property {number} freeMB
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
	* Megabytes to Gigabytes conversion
	* @param {number} mb
	* @returns {string}
	*/
	static toGB(mb) {
		return (mb / 1000).toFixed(2)
	}

	/**
	* Downloads an SVG file and returns its content as a string
	* @param {string} url
	* @returns {Promise<string>}
	*/
	static async inlineSVG(url) {
		const res = await fetch(url)
		return await res.text()
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
			const res = await fetch(endpoint)
			const data = await res.json()
			const formatted = await formatter(data)

			const el = document.getElementById(elementId)
			el.innerHTML = formatted

			const section = el.closest('section')
			if (section) {
				section.classList.remove('loading')
				section.classList.add('loaded')
			}
		} catch (err) {
			const el = document.getElementById(elementId)
			el.innerHTML = `<p class="error">Error loading data</p>`
			console.error(err)
		}
	}

	/**
	* Formats the OS data into HTML
	* @param {OSData} data
	* @returns {Promise<string>}
	*/
	static async formatOS(data) {
		const iconUrl = `https://raw.githubusercontent.com/lukas-w/font-logos/refs/heads/master/vectors/${data.id}.svg`
		const svg = await DoveDashUI.inlineSVG(iconUrl)
		return DoveDashUI.dedent(`
		<div class="os-block">
			<div class="os-header">
				<div class="os-icon">${svg}</div>
				<span class="os-name">${data.os}</span>
			</div>
			<ul class="os-list">
				<li><strong>Architecture:</strong> ${data.arch}</li>
				<li><strong>Kernel:</strong> ${data.kernel}</li>
				<li><strong>Uptime:</strong> ${data.uptime}</li>
				<li><strong>Hostname:</strong> ${data.hostname}</li>
			</ul>
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
		<ul class="cpu-list">
			<li><strong>Brand:</strong> ${data.brand}</li>
			<li><strong>Model:</strong> ${data.model}</li>
			<li><strong>Cores:</strong> ${data.cores}</li>
			<li><strong>Threads:</strong> ${data.threads}</li>
			<li><strong>Frequency:</strong> ${data.frequency} GHz</li>
		</ul>
		`)
	}

	/**
	* Formats the memory data into HTML
	* @param {MemData} data
	* @returns {string}
	*/
	static formatMem(data) {
		if (!data) return `<p class="mem-error">No memory data</p>`

		const used = isFinite(data.usedMB) ? DoveDashUI.toGB(data.usedMB) : '?'
		const total = isFinite(data.totalMB) ? DoveDashUI.toGB(data.totalMB) : '?'
		const percent = isFinite(data.usedPercent) ? data.usedPercent.toFixed(0) : '?'

		return `<p class="mem-line"><strong>Usage:</strong> ${used} GB / ${total} GB (${percent}%)</p>`
	}

	/**
	* Formats the storage data into HTML 
	* @param {StorageData} data
	* @returns {string}
	*/
	static formatStorage(data) {
		if (!data) return `<p class="storage-error">No storage data</p>`

		const used = isFinite(data.usedMB) ? DoveDashUI.toGB(data.usedMB) : '?'
		const total = isFinite(data.totalMB) ? DoveDashUI.toGB(data.totalMB) : '?'
		const percent = isFinite(data.usedPercent) ? data.usedPercent.toFixed(2) : '?'
		const mount = data.mountpoint || '/'
		const fs = data.fsType || 'unknown'

		return `<p class="storage-line"><strong>Disk (${mount}):</strong> ${used} GB / ${total} GB (${percent}%) – ${fs}</p>`
	}

	/**
	* Formats the sensor data into HTML
	* @param {SensorChip[]} data
	* @returns {string}
	*/
	static formatSensors(data) {
		return data.map(chip => DoveDashUI.dedent(`
		<div class="sensor-block">
			<h3 class="sensor-name">${chip.name}</h3>
			<p class="sensor-adapter"><strong>Adapter:</strong> ${chip.adapter}</p>
			<ul class="sensor-list">
				${chip.readings.map(r => `<li><strong>${r.label}:</strong> ${r.value.toFixed(1)} ${r.unit || ''} ${r.extra || ''}</li>`).join('')}
			</ul>
		</div>
		`)).join('')
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
	}
}

DoveDashUI.refreshAll()
setInterval(() => DoveDashUI.refreshAll(), 5000)
