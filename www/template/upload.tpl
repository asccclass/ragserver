<!DOCTYPE html>
<html lang="zh_TW">
<head>  
   <meta charset="UTF-8">
   <meta name="viewport" content="width=device-width, initial-scale=1.0">
   <link href='https://unpkg.com/boxicons@2.1.4/css/boxicons.min.css' rel='stylesheet'>
   <link href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" rel="stylesheet">
   <link rel="stylesheet" href="css/xx.css">
   <link rel="stylesheet" href="css/table.css">
   <title>Job Queue File Upload</title>
   <style>
      .right-align {
         text-align: right;
      }
   </style>
</head>
<body>
<!-- Sidebar -->
<div class="sidebar">
   <a href="#" class="logo">
      <i class='bx bx-cart-download'></i>
      <div class="logo-name"><span></span>Job Queue</div>
   </a>
   <ul class="side-menu">
       <li><a href="#"><i class='bx bxs-dashboard'></i>Home</a></li>
       <li class="active"><a href="/homepage"><i class='bx bx-analyse'></i>Upload Job Queue</a></li>
        <!--
          <li><a href="#"><i class='bx bx-store-alt'></i>System</a></li>
            <li><a href="#"><i class='bx bx-message-square-dots'></i>Tickets</a></li>
            <li><a href="#"><i class='bx bx-group'></i>Users</a></li>
            <li><a href="#"><i class='bx bx-cog'></i>Settings</a></li>
       -->
   </ul>
   <ul class="side-menu">
       <li>
           <a href="/logout" class="logout"><i class='bx bx-log-out-circle'></i>Logout</a>
       </li>
   </ul>
</div>
<!-- End of Sidebar -->
<!-- Main Content -->
<div class="content">
   <div class="right-align">您好，{{ .Userinfo.ChName }}（{{ .Userinfo.Email }}）</div>
   <div class="container">
      <div class="row justify-content-center mt-5">
         <div class="col-md-6">
            <div class="card">
               <div class="card-header">
                  請選擇聲音檔上傳，格式限制：mp3、wav
               </div>
               <div class="card-body">
                  <form id="uploadForm" enctype="multipart/form-data">
                     <div class="form-group">
                        <label for="file">Select MP3 File</label>
                        <input type="file" class="form-control-file" id="file" name="file" accept=".mp3">
                        <small id="fileHelp" class="form-text text-muted">Please upload an MP3 file.</small>
                     </div>
                     <button type="button" class="btn btn-primary" onclick="uploadFile()">Upload</button>
                  </form>
                  <div class="progress mt-3" style="display:none;">
                        <div id="progressBar" class="progress-bar" role="progressbar" style="width: 0%;" aria-valuenow="0" aria-valuemin="0" aria-valuemax="100">0%</div>
                  </div>
                  <div id="uploadStatus" class="mt-3"></div>
               </div>
            </div>
            <main>
               <div class="table-container">
                  {{ $length := len .Jobs }} 
                  {{ if eq $length 0 }}
                  <table id="data-table">
                     <thead>
                        <tr>
                           <th>檔案名稱</th>
                           <th>工作型態</th>
                           <th>目前狀態</th>
                           <th>上傳時間</th>
                           <th>結束時間</th>
                           <!-- <th>下載</th> -->
                        </tr>
                     </thead>
                     <tbody id="table-body">
                        {{range .Jobs}}
                        <tr>
                           <td>{{ .FileName }}</td>
                           <td>{{ .Action }}</td>
                           <td>{{ .Status }}</td>
                           <td>{{ .UploadTime }}</td>
                           <td>{{ .FinishTime }}&nbsp;</td>
                           <!-- <td><a href="/download/{{- .FileName -}}" target=_blank>下載</a></td> -->
                        </tr>
                        {{end}}
                     </tbody>
                  </table>
                  {{ end }}
               </div>
            </main>
         </div>
      </div>
   </div>
</div>
<script src="https://code.jquery.com/jquery-3.5.1.slim.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/@popperjs/core@2.5.3/dist/umd/popper.min.js"></script>
<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.min.js"></script>
<script>
   function generateTable(data) {
      console.log(data);
      var tableBody = document.getElementById('table-body');
      tableBody.innerHTML = '';
      data.forEach(function(item) {
          var row = document.createElement('tr');
          var nameCell = document.createElement('td');
          nameCell.textContent = item.fileName;
          row.appendChild(nameCell);

          var ageCell = document.createElement('td');
          ageCell.textContent = item.action;
          row.appendChild(ageCell);

          var ageCell = document.createElement('td');
          ageCell.textContent = item.status;
          row.appendChild(ageCell);

          var ageCell = document.createElement('td');
          ageCell.textContent = item.uploadTime;
          row.appendChild(ageCell);

          tableBody.appendChild(row);
      });
   }

   function generateList(data) {
      console.log(data);
      var listContainer = document.getElementById('jobLists');
      listContainer.innerHTML = '';
      data.forEach(function(item) {
         var listItem = document.createElement('li');
         listItem.textContent = item.name; // 假設JSON資料中有一個名為'name'的屬性
         listContainer.appendChild(listItem); // 將li元素添加到列表容器中
      });
   }

   function uploadFile() {
      var fileInput = document.getElementById('file');
      var file = fileInput.files[0];
      var formData = new FormData();
      formData.append('file', file);
      var xhr = new XMLHttpRequest();
      xhr.open('POST', '/upload', true);
      xhr.upload.onprogress = function (e) {
         if(e.lengthComputable) {
            var progress = (e.loaded / e.total) * 100;
            document.getElementById('progressBar').style.width = progress + '%';
            document.getElementById('progressBar').innerText = Math.round(progress) + '%';
         }
      };
      xhr.onload = function () {
         if(xhr.status === 200) {
            var responseData = JSON.parse(xhr.responseText);
            generateTable(responseData);
            // generateList(responseData);
         } else {
            document.getElementById('uploadStatus').innerText = 'Upload failed';
         }
         document.getElementById('progressBar').style.width = '0%';
         document.getElementById('progressBar').innerText = '0%';
         document.getElementById('uploadForm').reset();
       };
       xhr.onerror = function () {
          document.getElementById('uploadStatus').innerText = 'Upload failed';
          document.getElementById('progressBar').style.width = '0%';
          document.getElementById('progressBar').innerText = '0%';
          document.getElementById('uploadForm').reset();
       };
       xhr.send(formData);
    }
   function validateForm() {
      var fileInput = document.getElementById('file');
      var filePath = fileInput.value;
      var allowedExtensions = /(\.wav)|(\.mp3)$/i;

      if(!allowedExtensions.exec(filePath)) {
          alert('Please upload an MP3 file.');
          fileInput.value = '';
          return false;
      }
      return true;
   }
</script>
</body>
</html>
