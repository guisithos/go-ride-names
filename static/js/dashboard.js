// Activity type icons mapping
const activityIcons = {
    Run: "🏃",
    Ride: "🚴",
    Swim: "🏊",
    Walk: "🚶",
    WeightTraining: "💪",
    Workout: "💪",
    Yoga: "🧘",
    CrossFit: "🏋️",
    VirtualRide: "🎮",
};

// Define stationary activities that should show duration instead of distance
const stationaryActivities = ['WeightTraining', 'Workout', 'Yoga', 'CrossFit'];

// Format duration in seconds to hours and minutes
function formatDuration(seconds) {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (hours > 0) {
        return `${hours}h${minutes}min`;
    }
    return `${minutes}min`;
}

// Store activities for later use
window.remainingActivities = [];

// Format distance in meters to km
function formatDistance(meters) {
    const km = meters / 1000;
    return `${km.toFixed(2)}km`;
}

// Format date to local string
function formatDate(dateStr) {
    return new Date(dateStr).toLocaleDateString('pt-BR', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

// Load activities from Strava
async function loadActivities() {
    try {
        const response = await fetch('https://www.strava.com/api/v3/athlete/activities', {
            headers: {
                'Authorization': `Bearer ${accessToken}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to fetch activities');
        }

        const activities = await response.json();
        
        // Process activities for stats
        const stats = processActivitiesStats(activities);
        displayStats(stats);

        // Display recent activities (first 5)
        displayRecentActivities(activities.slice(0, 5));

        // Store remaining activities
        window.remainingActivities = activities.slice(5);

    } catch (error) {
        console.error('Error:', error);
    }
}

// Update processActivitiesStats to handle both distance and duration
function processActivitiesStats(activities) {
    const stats = {};
    
    activities.forEach(activity => {
        if (!stats[activity.type]) {
            stats[activity.type] = {
                count: 0,
                distance: 0,
                duration: 0
            };
        }
        stats[activity.type].count++;
        stats[activity.type].distance += activity.distance;
        stats[activity.type].duration += activity.moving_time; // Add duration tracking
    });

    return stats;
}

// Update displayStats to show either distance or duration based on activity type
function displayStats(stats) {
    const container = document.getElementById('activity-stats');
    container.innerHTML = '';
    
    Object.entries(stats).forEach(([type, data]) => {
        const div = document.createElement('div');
        div.className = 'activity-stat';
        const isStationary = stationaryActivities.includes(type);
        
        div.innerHTML = `
            <div class="icon">${activityIcons[type] || '🏃'}</div>
            <div class="count">${data.count} ${type}</div>
            <div class="metric">
                ${isStationary 
                    ? `Tempo: ${formatDuration(data.duration)}`
                    : `Distância: ${formatDistance(data.distance)}`
                }
            </div>
        `;
        container.appendChild(div);
    });
}

// Update displayRecentActivities to show appropriate metrics
function displayRecentActivities(activities) {
    const container = document.getElementById('activities-recent');
    container.innerHTML = '';
    
    activities.forEach(activity => {
        const div = document.createElement('div');
        div.className = 'activity';
        const isStationary = stationaryActivities.includes(activity.type);
        
        div.innerHTML = `
            <h3>${activity.name}</h3>
            <p>Tipo: ${activity.type}</p>
            ${isStationary 
                ? `<p>Tempo: ${formatDuration(activity.moving_time)}</p>`
                : `<p>Distância: ${formatDistance(activity.distance)}</p>`
            }
            <p>Data: ${formatDate(activity.start_date_local)}</p>
        `;
        container.appendChild(div);
    });
}

// Update toggleMoreActivities to show appropriate metrics
function toggleMoreActivities() {
    const expandedContainer = document.getElementById('activities-expanded');
    const button = document.querySelector('.show-more-btn');
    
    if (expandedContainer.classList.contains('show')) {
        expandedContainer.classList.remove('show');
        button.textContent = 'Mostrar mais atividades';
    } else {
        expandedContainer.classList.add('show');
        button.textContent = 'Mostrar menos';
        
        expandedContainer.innerHTML = '';
        window.remainingActivities.forEach(activity => {
            const div = document.createElement('div');
            div.className = 'activity';
            const isStationary = stationaryActivities.includes(activity.type);
            
            div.innerHTML = `
                <h3>${activity.name}</h3>
                <p>Tipo: ${activity.type}</p>
                ${isStationary 
                    ? `<p>Tempo: ${formatDuration(activity.moving_time)}</p>`
                    : `<p>Distância: ${formatDistance(activity.distance)}</p>`
                }
                <p>Data: ${formatDate(activity.start_date_local)}</p>
            `;
            expandedContainer.appendChild(div);
        });
    }
}

// Add event listeners for buttons
document.getElementById('rename').addEventListener('click', async function() {
    const button = this;
    button.disabled = true;
    button.innerHTML = '<span>Renomeando...</span>';
    
    try {
        const response = await fetch('/rename-activities', {
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error('Failed to rename activities');
        }

        loadActivities(); // Reload activities after renaming
    } catch (error) {
        console.error('Error:', error);
    } finally {
        button.disabled = false;
        button.innerHTML = '<span>Renomear Todas</span>';
    }
});

document.getElementById('subscribe').addEventListener('click', async function() {
    const button = this;
    button.disabled = true;
    button.innerHTML = '<span>Ativando...</span>';
    
    try {
        const response = await fetch('/subscribe', {
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error('Failed to subscribe');
        }

        checkSubscriptionStatus();
    } catch (error) {
        console.error('Error:', error);
    } finally {
        button.disabled = false;
        button.innerHTML = '<span>Ativar Auto-Renomeação</span>';
    }
});

// Function to check subscription status
async function checkSubscriptionStatus() {
    try {
        const response = await fetch('/subscription-status');
        const data = await response.json();
        
        const statusDiv = document.getElementById('subscription-status');
        if (data.active) {
            statusDiv.className = 'status active';
            statusDiv.textContent = 'Auto-renomeação está ativa';
        } else {
            statusDiv.className = 'status inactive';
            statusDiv.textContent = 'Auto-renomeação está inativa';
        }
    } catch (error) {
        console.error('Error checking status:', error);
    }
}

// Check status and load activities when page loads
checkSubscriptionStatus();
loadActivities();