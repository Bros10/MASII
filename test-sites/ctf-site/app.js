var apiUrl = "https://127.0.0.1:5000/secret-endpoint";

// Make request to the hidden endpoint
fetch(apiUrl)
  .then(response => {
    console.log(response);
  })
  .catch(error => {
    console.log(error);
  });
