<script lang="ts">
    export let value:string
    export let placeholder:string
    export let label:string
    export let name:string
    export let id:string
</script>

<td><label for={id}>{label}</label></td>
<td><input bind:value {placeholder} {id} {name} type="text"/></td>
