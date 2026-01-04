// js/profiles.js

const tableBody = document.getElementById('profiles-table-body');
const mikrotikFilter = document.getElementById('mikrotik-filter');
const syncBtn = document.getElementById('sync-btn');

// Fetch Routers for Filter
async function fetchRouters() {
    try {
        const res = await fetch(`${API_BASE}/api/mikrotiks`);
        const result = await res.json();
        if (result.status === 'success') {
            result.data.forEach(mk => {
                const option = document.createElement('option');
                option.value = mk.id;
                option.textContent = mk.name;
                mikrotikFilter.appendChild(option);
            });
        }
    } catch (err) {
        console.error('Error fetching routers:', err);
    }
}

// Fetch Profiles
async function fetchProfiles() {
    tableBody.innerHTML = `<tr><td colspan="6" style="text-align: center; padding: 20px;"><i class="fa-solid fa-spinner fa-spin"></i> Loading...</td></tr>`;

    let url = `${API_BASE}/api/profiles?limit=100`;
    if (mikrotikFilter.value) {
        url += `&mikrotik_id=${mikrotikFilter.value}`;
    }

    try {
        const res = await fetch(url);
        const result = await res.json();

        if (result.status === 'success') {
            renderTable(result.data);
        } else {
            tableBody.innerHTML = `<tr><td colspan="6" style="text-align: center; color: red;">Error: ${result.message}</td></tr>`;
        }
    } catch (err) {
        console.error('Error fetching profiles:', err);
        tableBody.innerHTML = `<tr><td colspan="6" style="text-align: center; color: red;">Failed to load profiles</td></tr>`;
    }
}

function renderTable(profiles) {
    if (profiles.length === 0) {
        tableBody.innerHTML = `<tr><td colspan="6" style="text-align: center; padding: 20px;">No profiles found.</td></tr>`;
        return;
    }

    tableBody.innerHTML = '';
    profiles.forEach(profile => {
        const row = document.createElement('tr');

        let rateLimit = '-';
        if (profile.rate_limit_up || profile.rate_limit_down) {
            rateLimit = `${profile.rate_limit_up || 'Unlim'} / ${profile.rate_limit_down || 'Unlim'}`;
        }

        row.innerHTML = `
            <td>
                <div style="font-weight: 600; color: var(--gray-900);">${profile.name}</div>
                <div style="font-size: 0.75rem; color: var(--gray-500);">${profile.mikrotik_id}</div>
            </td>
            <td><span class="status-badge status-active" style="text-transform: uppercase;">${profile.profile_type}</span></td>
            <td>${rateLimit}</td>
            <td>${profile.pppoe?.remote_address || '-'}</td>
            <td>Rp ${profile.price ? profile.price.toLocaleString() : '0'}</td>
            <td>
                <button class="btn btn-sm btn-secondary" onclick="syncProfile('${profile.id}')" title="Sync to MikroTik">
                    <i class="fa-solid fa-sync"></i>
                </button>
                <button class="btn btn-sm btn-secondary" onclick="deleteProfile('${profile.id}')" title="Delete">
                    <i class="fa-solid fa-trash" style="color: #ef4444;"></i>
                </button>
            </td>
        `;
        tableBody.appendChild(row);
    });
}

// Sync Profile
window.syncProfile = async (id) => {
    if (!confirm('Sync this profile to MikroTik?')) return;
    try {
        const res = await fetch(`${API_BASE}/api/profiles/${id}/sync`, { method: 'POST' });
        const result = await res.json();
        if (res.ok) {
            alert('Synced successfully');
        } else {
            alert('Sync failed: ' + result.message);
        }
    } catch (err) {
        alert('Sync error');
    }
};

// Delete Profile
window.deleteProfile = async (id) => {
    if (!confirm('Are you sure you want to delete this profile?')) return;
    try {
        const res = await fetch(`${API_BASE}/api/profiles/${id}`, { method: 'DELETE' });
        if (res.ok) {
            fetchProfiles();
        } else {
            alert('Failed to delete');
        }
    } catch (err) {
        alert('Error deleting');
    }
};

// Event Listeners
mikrotikFilter.addEventListener('change', fetchProfiles);

syncBtn.addEventListener('click', async () => {
    const mikrotikId = mikrotikFilter.value;
    if (!mikrotikId) {
        alert('Please select a router first to sync all profiles.');
        return;
    }
    if (!confirm('Sync ALL profiles from MikroTik? This might overwrite local changes.')) return;

    try {
        await fetch(`${API_BASE}/api/profiles/sync-all/${mikrotikId}`, { method: 'POST' });
        alert('Sync started/completed.');
        fetchProfiles();
    } catch (err) {
        alert('Sync failed');
    }
});

// Init
fetchRouters();
fetchProfiles();
