#include <stdlib.h>
#include <stdio.h>

// Check if CUDA is available by conditionally including CUDA headers
#if defined(_WIN32)
#include <windows.h>
#include <cfgmgr32.h>
#ifdef __CUDA_RUNTIME_H__
#include <cuda_runtime.h>
#endif
#elif defined(__linux__)
#ifdef __CUDA_RUNTIME_H__
#include <cuda_runtime.h>
#endif
#endif

// GPU memory detection function
int get_gpu_memory_info(int device_id, size_t* free_memory, size_t* total_memory) {
    #ifdef __CUDA_RUNTIME_H__
    // CUDA is available
    if (cudaSetDevice(device_id) != cudaSuccess) {
        return -1; // Device not available
    }
    
    if (cudaMemGetInfo(free_memory, total_memory) != cudaSuccess) {
        return -2; // Failed to get memory info
    }
    
    return 0; // Success
    #else
    // No CUDA support - try platform-specific alternatives
    #if defined(_WIN32)
    // On Windows, query using WMI or other methods
    *free_memory = 0;
    *total_memory = 1024 * 1024 * 1024; // Assume 1GB as placeholder
    return -3;
    #elif defined(__linux__) && defined(__GNUC__)
    // On Linux, try to read from /proc/meminfo or similar
    *free_memory = 0;
    *total_memory = 1024 * 1024 * 1024; // Assume 1GB as placeholder
    return -3;
    #else
    // Generic fallback
    *free_memory = 0;
    *total_memory = 0;
    return -3; // No CUDA support
    #endif
    #endif
}
