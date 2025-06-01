<script>
  import { onMount } from 'svelte';
  
  let count = 0;
  let mounted = false;
  
  onMount(() => {
    mounted = true;
    console.log('Test page mounted successfully');
  });
  
  function increment() {
    count++;
  }
</script>

<h1>Development Test Page</h1>
<p>This is a simple test page without any authentication logic.</p>
<p>If you can see this, the /dev/ route is working correctly!</p>

<div class="test-section">
  <h2>JavaScript Test</h2>
  <p>Mounted: {mounted ? 'Yes' : 'No'}</p>
  <p>Count: {count}</p>
  <button onclick={increment}>Click me to test JavaScript</button>
</div>

<style>
  h1 { color: #ff3e00; }
  p { margin: 1rem 0; }
  .test-section { 
    border: 2px solid #ff3e00; 
    padding: 1rem; 
    margin: 1rem 0; 
  }
  button {
    background: #ff3e00;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    cursor: pointer;
    border-radius: 4px;
  }
</style>
