function deleteUser(userId) {
    if (confirm('Bu kullanıcıyı silmek istediğinizden emin misiniz? Bu işlem geri alınamaz!')) {
        fetch(`/web-delete?id=${userId}`, {
            method: 'GET'
        })
        .then(response => {
            if (response.ok) {
                location.reload();
            } else {
                alert('Kullanıcı silinirken hata oluştu.');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Kullanıcı silinirken hata oluştu.');
        });
    }
}

function editUser(userId, name, surname, email, age, phone, role, photoPath) {
    // Modal formunu doldur
    document.getElementById('editUserId').value = userId;
    document.getElementById('editModalName').value = name;
    document.getElementById('editModalSurname').value = surname;
    document.getElementById('editModalEmail').value = email;
    document.getElementById('editModalAge').value = age;
    document.getElementById('editModalPhone').value = phone;
    document.getElementById('editModalRole').value = role;
    document.getElementById('editModalPassword').value = '';
    
    // Mevcut fotoğrafı göster
    const photoPreview = document.getElementById('currentPhotoPreview');
    if (photoPath) {
        photoPreview.innerHTML = `<img src="${photoPath}" alt="Mevcut fotoğraf" style="width: 100px; height: 100px; object-fit: cover; border-radius: 8px;">`;
    } else {
        photoPreview.innerHTML = '<p class="text-muted">Mevcut fotoğraf yok</p>';
    }
    
    // Modal'ı aç
    const modal = new bootstrap.Modal(document.getElementById('editUserModal'));
    modal.show();
}

document.addEventListener('DOMContentLoaded', function() {
    // Auto-hide alerts after 5 seconds
    const alerts = document.querySelectorAll('.alert');
    alerts.forEach(alert => {
        setTimeout(() => {
            const bsAlert = new bootstrap.Alert(alert);
            bsAlert.close();
        }, 5000);
    });
});
