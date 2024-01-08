import pandas as pd

not_web_related_issues = [
    "SMB Signing not required",
    "Network Services Accessible",
    "Insecure Android OS Supported",
    "Jailbreak / Rooted Device Detection Not Implemented",
    "LLMNR/NBT-NS Enabled",
    "Remote Desktop Protocol (RDP) Security Issues",
    "IPv6 Supported but not in use (DNS Poisoning)",
    "Simple Network Management Protocol (SNMP) Default Community Name",
    "Unsupported Windows versions in use",
    "Lucky Thirteen attack",
    "SSH Security Issues",
    "Printers with default admin credentials",
    "Unencrypted Telnet Server",
    "Internet Key Exchange (IKE) Aggressive Mode with Pre-Shared Key",
    "Microsoft Windows SMBv1 Enabled",
    "Credentials Exposure",
    "Management/Dangerous Services Exposed to the Internet",
    "SSH Weak Algorithms Supported",
    "Domain Users in Local Administrators group",
    "SMTP User/Email Enumeration",
    "NTP Amplification Attack",
    "VNC Service Accessible Without Authentication",
    "Clear Text Protocols in Use (FTP)",
    "Unsupported Web Server Detection",
    "Unencrypted FTP Service",
    "SSH Weak Key Exchange Algorithms Detected",
    "SMTP Service Cleartext Login Permitted",
    "Unencrypted Telnet Service",
    "Screenshots Not Disabled",
    "OpenSSL ‘CHANGECIPHERSPEC’ MiTM Vulnerability",
    "Unprotected Telnet Service"
]

input_file = "final-Issue-Numbers.csv"
output_file = "filtered_blockusage_without_nonweb.csv"

# Read the CSV file
data = pd.read_csv(input_file)

# Filter out the not web-related issues
filtered_data = data[~data["name"].isin(not_web_related_issues)]

# Save the filtered data to a new CSV file
filtered_data.to_csv(output_file, index=False)
