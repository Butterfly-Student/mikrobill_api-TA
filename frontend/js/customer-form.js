// Customer form logic for Add Customer page

const form = document.getElementById('add-customer-form');
const typeSelect = document.getElementById('service-type-select');
const pppoeFields = document.getElementById('pppoe-fields');
const mikrotikSelect = document.getElementById('mikrotik-select');
const profileSelect = document.getElementById('pppoe-profile-select');

// Fetch MikroTik Routers on load
async function fetchMikrotiks() {
    try {
        const res = await fetch(`${API_BASE}/api/mikrotiks`);
        const result = await res.json();
        if (result.status === 'success') {
            mikrotikSelect.innerHTML = '<option value="">Select Router</option>';
            result.data.forEach(mk => {
                const option = document.createElement('option');
                option.value = mk.id;
                option.textContent = `${mk.name} (${mk.host})`;
                mikrotikSelect.appendChild(option);
            });
        }
    } catch (err) {
        console.error('Failed to fetch mikrotiks:', err);
    }
}

// Fetch Profiles when Mikrotik is selected
async function fetchProfiles(mikrotikId) {
    if (!mikrotikId) {
        profileSelect.innerHTML = '<option value="">Select Profile</option>';
        return;
    }

    try {
        const res = await fetch(`${API_BASE}/api/profiles?mikrotik_id=${mikrotikId}&limit=100`);
        const result = await res.json();
        if (result.status === 'success') {
            profileSelect.innerHTML = '<option value="">Select Profile</option>';
            result.data.forEach(profile => {
                if (profile.profile_type === 'pppoe') {
                    const option = document.createElement('option');
                    option.value = profile.id; // Use ID, backend expects ID
                    option.textContent = `${profile.name} (${profile.rate_limit_up || 'Unlim'}/${profile.rate_limit_down || 'Unlim'})`;
                    profileSelect.appendChild(option);
                }
            });
        }
    } catch (err) {
        console.error('Failed to fetch profiles:', err);
    }
}

mikrotikSelect.addEventListener('change', (e) => {
    fetchProfiles(e.target.value);
});

typeSelect.addEventListener('change', () => {
    if (typeSelect.value === 'pppoe') {
        pppoeFields.style.display = 'block';
    } else {
        pppoeFields.style.display = 'none';
        profileSelect.value = "";
    }
});

form.addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = new FormData(form);
    const data = Object.fromEntries(formData.entries());

    const payload = {
        mikrotik_id: data.mikrotik_id,
        name: data.name,
        username: data.username,
        service_type: data.service_type,
        phone: data.phone || null,
        email: data.email || null,
        pppoe_username: data.pppoe_username || null,
        pppoe_password: data.pppoe_password || null,
        pppoe_profile_id: data.pppoe_profile_id || null
    };

    try {
        const res = await fetch(`${API_BASE}/api/customers`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });

        const result = await res.json();

        if (res.ok && result.status === 'success') {
            alert('Customer created successfully!');
            window.location.href = '/';
        } else {
            alert('Error: ' + (result.message || 'Unknown error'));
        }
    } catch (err) {
        console.error(err);
        alert('Failed to submit form');
    }
});

// Initialize
fetchMikrotiks();