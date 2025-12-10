`lscpu`


```sql

Architecture:             x86_64
  CPU op-mode(s):         32-bit, 64-bit
  Address sizes:          43 bits physical, 48 bits virtual
  Byte Order:             Little Endian
CPU(s):                   4
  On-line CPU(s) list:    0-3
Vendor ID:                AuthenticAMD
  Model name:             AMD Ryzen 3 2200G with Radeon Vega Graphics
    CPU family:           23
    Model:                17
    Thread(s) per core:   1
    Core(s) per socket:   4
    Socket(s):            1
    Stepping:             0
    Frequency boost:      enabled
    CPU max MHz:          3500,0000
    CPU min MHz:          1600,0000
    BogoMIPS:             6986.59
    Flags:                fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 c
                          lflush mmx fxsr sse sse2 ht syscall nx mmxext fxsr_opt pdpe1gb rdtscp lm 
                          constant_tsc rep_good nopl nonstop_tsc cpuid extd_apicid aperfmperf rapl 
                          pni pclmulqdq monitor ssse3 fma cx16 sse4_1 sse4_2 movbe popcnt aes xsave
                           avx f16c rdrand lahf_lm cmp_legacy svm extapic cr8_legacy abm sse4a misa
                          lignsse 3dnowprefetch osvw skinit wdt tce topoext perfctr_core perfctr_nb
                           bpext perfctr_llc mwaitx cpb hw_pstate ssbd ibpb vmmcall fsgsbase bmi1 a
                          vx2 smep bmi2 rdseed adx smap clflushopt sha_ni xsaveopt xsavec xgetbv1 c
                          lzero irperf xsaveerptr arat npt lbrv svm_lock nrip_save tsc_scale vmcb_c
                          lean flushbyasid decodeassists pausefilter pfthreshold avic v_vmsave_vmlo
                          ad vgif overflow_recov succor smca sev sev_es ibpb_exit_to_user
Virtualization features:  
  Virtualization:         AMD-V
Caches (sum of all):      
  L1d:                    128 KiB (4 instances)
  L1i:                    256 KiB (4 instances)
  L2:                     2 MiB (4 instances)
  L3:                     4 MiB (1 instance)
NUMA:                     
  NUMA node(s):           1
  NUMA node0 CPU(s):      0-3
Vulnerabilities:          
  Gather data sampling:   Not affected
  Itlb multihit:          Not affected
  L1tf:                   Not affected
  Mds:                    Not affected
  Meltdown:               Not affected
  Mmio stale data:        Not affected
  Reg file data sampling: Not affected
  Retbleed:               Mitigation; untrained return thunk; SMT disabled
  Spec rstack overflow:   Mitigation; SMT disabled
  Spec store bypass:      Mitigation; Speculative Store Bypass disabled via prctl
  Spectre v1:             Mitigation; usercopy/swapgs barriers and __user pointer sanitization
  Spectre v2:             Mitigation; Retpolines; IBPB conditional; STIBP disabled; RSB filling; PB
                          RSB-eIBRS Not affected; BHI Not affected
  Srbds:                  Not affected
  Tsx async abort:        Not affected
  Vmscape:                Mitigation; IBPB before exit to userspace

```

  `df -h --total`


```sql
No LSB modules are available.
Distributor ID: Ubuntu
Description:    Ubuntu 22.04.5 LTS
Release:        22.04
Codename:       jammy
vps1@vps1:~/WorkSpace/deepapp_golang_grpc_hub/services/web-api$ df -h --total
Filesystem      Size  Used Avail Use% Mounted on
tmpfs           1,5G  2,5M  1,5G   1% /run
/dev/nvme0n1p2  234G   94G  128G  43% /
tmpfs           7,3G     0  7,3G   0% /dev/shm
tmpfs           5,0M  4,0K  5,0M   1% /run/lock
efivarfs        150K   80K   66K  55% /sys/firmware/efi/efivars
/dev/nvme0n1p1  511M  6,1M  505M   2% /boot/efi
tmpfs           1,5G  132K  1,5G   1% /run/user/1000
total           244G   94G  139G  41% 

```