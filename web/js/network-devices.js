import { gosightFetch } from "./api.js";

// State management
let devices = [];
let currentPage = 0;
const pageSize = 10;

// DOM Elements
const tableBody = document.getElementById('network-device-table-body');
const filterInput = document.getElementById('filter-by');
const filterStatus = document.getElementById('filter-status');
const addDeviceBtn = document.createElement('button');

// Initialize the page
async function init() {

    setupModal();
    setupFilters();
    await loadDevices();
}

// Show Edit Device Modal
window.showEditDeviceModal = function(deviceId) {
    const device = devices.find(d => d.ID === deviceId);
    if (!device) {
        showToast('Device not found', 'failure');
        return;
    }
    // Populate the form with device data
    const form = document.getElementById('add-device-form');
    form.elements['name'].value = device.Name || '';
    form.elements['vendor'].value = device.Vendor || '';
    form.elements['address'].value = device.Address || '';
    form.elements['protocol'].value = device.Protocol ? device.Protocol.toLowerCase() : '';
    form.elements['format'].value = device.Format || '';
    form.elements['facility'].value = device.Facility || '';
    form.elements['syslog_id'].value = device.SyslogID || '';
    form.elements['rate_limit'].value = device.RateLimit || '';
    // Store editing state
    form.setAttribute('data-edit-id', deviceId);
    // Show the modal
    document.querySelector('[data-modal-target="add-device-modal"]').click();
}

// Setup the modal
function setupModal() {
    // Handle form submission
    document.getElementById('add-device-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const form = e.target;
        const formData = new FormData(form);
        const device = Object.fromEntries(formData.entries());

        // Basic validation (use lowercase keys)
        if (!device.name || !device.address || !device.protocol) {
            showToast('Name, Address, and Protocol are required.', 'failure');
            return;
        }

        // Sanitize
        device.name = device.name.trim();
        device.vendor = device.vendor ? device.vendor.trim() : '';
        device.address = device.address.trim();
        device.protocol = device.protocol.trim().toUpperCase(); // Ensure uppercase for backend
        device.format = device.format ? device.format.trim() : '';
        device.facility = device.facility ? device.facility.trim() : '';
        // Fix: use correct keys for backend (SyslogID, RateLimit)
        device.SyslogID = device.syslog_id ? device.syslog_id.trim() : '';
        device.RateLimit = device.rate_limit ? parseInt(device.rate_limit, 10) : 0;
        delete device.syslog_id;
        delete device.rate_limit;

        // If editing, send PUT to /api/v1/network-devices/{id}
        const editId = form.getAttribute('data-edit-id');
        let url = '/api/v1/network-devices';
        let method = 'POST';
        if (editId) {
            url = `/api/v1/network-devices/${editId}`;
            method = 'PUT';
        }

        try {
            const response = await gosightFetch(url, {
                method,
                body: JSON.stringify(device)
            });
            if (!response.ok) {
                const errorText = await response.text();
                showToast('Failed to save device: ' + errorText, 'failure');
                return;
            }
            document.querySelector('[data-modal-hide="add-device-modal"]').click();
            showToast('Device saved successfully', 'success');
            await loadDevices();
        } catch (error) {
            console.error('Failed to save device:', error);
            showToast('Failed to save device', 'failure');
        } finally {
            form.removeAttribute('data-edit-id');
        }
    });
}

// Setup filters
function setupFilters() {
    filterInput.addEventListener('input', debounce(renderDevices, 300));
    filterStatus.addEventListener('change', renderDevices);
}

// Load devices from API
async function loadDevices() {
    try {
        const response = await gosightFetch(
            `/api/v1/network-devices?offset=${currentPage * pageSize}&limit=${pageSize}`
        );
        
        // First read the response as JSON
        const data = await response.json();
        console.log('⟢ parsed JSON data:', data);

        // Try accessing the data with different possible structures
        devices = Array.isArray(data.devices) ? data.devices : 
                 Array.isArray(data) ? data : [];
        
        console.log('⟢ final devices array (length=', devices.length, '):', devices);

        renderDevices();
    } catch (error) {
        console.error('Failed to load devices:', error);
        showToast('Failed to load devices', 'failure');
    }
}

// Render devices in the table
function renderDevices() {
    tableBody.innerHTML = '';

    const query = filterInput.value.toLowerCase();
    const statusFilter = filterStatus.value;

    const filteredDevices = devices.filter(device => {
        // Match status if a filter is selected
        const matchesStatus = !statusFilter || device.Status === statusFilter;

        // Combine all searchable fields into one string
        const searchTarget = [
            device.Name,
            device.Vendor,
            device.Address,
            device.Port,
            device.Protocol,
            device.Format,
            device.Facility,
            device.SyslogID
        ].join(' ').toLowerCase();

        const matchesQuery = searchTarget.includes(query);

        return matchesStatus && matchesQuery;
    });

    if (filteredDevices.length === 0) {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td colspan="11" class="px-6 py-4 text-center text-gray-500 dark:text-gray-400">
                No network devices found
            </td>
        `;
        tableBody.appendChild(row);
        return;
    }

    filteredDevices.forEach(device => {
        const row = document.createElement('tr');
        row.className = 'bg-white border-b dark:bg-gray-800 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600';

        const isEnabled = device.Status === 'enabled';
        const statusPill = isEnabled
            ? `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                   <i class="fas fa-circle text-green-500 mr-1"></i>Enabled
               </span>`
            : `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
                   <i class="fas fa-circle text-red-500 mr-1"></i>Disabled
               </span>`;

        row.innerHTML = `
            <td class="px-6 py-4">${statusPill}</td>
            <td class="px-6 py-4 font-medium text-gray-900 whitespace-nowrap dark:text-white">${device.Name}</td>
            <td class="px-6 py-4 text-gray-500 dark:text-gray-400">${device.Vendor}</td>
            <td class="px-6 py-4 text-gray-500 dark:text-gray-400">${device.Address}</td>
            <td class="px-6 py-4 text-gray-500 dark:text-gray-400">${device.Port}</td>
            <td class="px-6 py-4 text-gray-500 dark:text-gray-400">${device.Protocol}</td>
            <td class="px-6 py-4 text-gray-500 dark:text-gray-400">${device.Format}</td>
            <td class="px-6 py-4 text-gray-500 dark:text-gray-400">${device.Facility}</td>
            <td class="px-6 py-4 text-gray-500 dark:text-gray-400">${device.SyslogID}</td>
            <td class="px-6 py-4 text-gray-500 dark:text-gray-400">${device.RateLimit}</td>
            <td class="px-6 py-4">
                <div class="flex space-x-2">
                    <button onclick="toggleDevice('${device.ID}')" class="font-medium text-blue-600 dark:text-blue-500 hover:underline">
                        <i class="fas fa-power-off"></i>
                    </button>
                    <button onclick="deleteDevice('${device.ID}')" class="font-medium text-red-600 dark:text-red-500 hover:underline">
                        <i class="fas fa-trash"></i>
                    </button>
                    <button onclick="showEditDeviceModal('${device.ID}')" class="font-medium text-yellow-600 dark:text-yellow-400 hover:underline">
                        <i class="fas fa-edit"></i>
                    </button>
                </div>
            </td>
        `;

        tableBody.appendChild(row);
    });
}
// Toggle device status
async function toggleDevice(id) {
    try {
        await gosightFetch(`/api/v1/network-devices/${id}/toggle`, {
            method: 'POST'
        });
        await loadDevices();
        showToast('Device status updated', 'success');
    } catch (error) {
        console.error('Failed to toggle device:', error);
        showToast('Failed to update device status', 'failure');
    }
}

// Delete device
async function deleteDevice(id) {
    if (!confirm('Are you sure you want to delete this device?')) return;

    try {
        await gosightFetch(`/api/v1/network-devices/${id}`, {
            method: 'DELETE'
        });
        await loadDevices();
        showToast('Device deleted successfully', 'success');
    } catch (error) {
        console.error('Failed to delete device:', error);
        showToast('Failed to delete device', 'failure');
    }
}

// Show toast notification
function showToast(message, type = 'success') {
    const toastElement = document.getElementById('toast');
    const messageElement = toastElement.querySelector('.toast-message');
    messageElement.textContent = message;

    // Optionally set color based on type
    toastElement.classList.remove('hidden');
    setTimeout(() => toastElement.classList.add('hidden'), 3000);
}

// Utility function for debouncing
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// Initialize the page when the DOM is loaded
document.addEventListener('DOMContentLoaded', init);