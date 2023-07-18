├─api                # REST API controllers
├─cli                # CLI application
├─client             # HTTP client
├─config             # Configuration loading and validation
├─contrib            
│ ├─ansible          # Ansible playbooks and roles
│ ├─docker-compose   # Docker Compose file for deployment
│ ├─grafana          # Grafana dashboard templates  
│ └─sysvinit         # Sysvinit scripts
│   └─etc  
│     ├─default      # Default configuration
│     └─init.d       # Init scripts
├─crypto             # Password hashing
├─database           # Database abstraction layer
├─errors             # Error types
├─fever              # Fever API implementation
├─googlereader       # Google Reader API client
├─http   
│ ├─client          # HTTP client
│ ├─cookie          # Cookie handling
│ ├─request         # Request handling
│ ├─response        # Response handling
│ ├─route           # Router using Gorilla mux
│ └─server          # HTTP server
├─integration        # Integration clients like Pocket and Wallabag
├─locale             # I18n translations
├─logger             # Loggers for console, file, syslog
├─metric             # Prometheus metrics
├─model              # Database models
├─oauth2             # OAuth2 support
├─packaging          # Packaging for Debian, Docker, RPM
├─proxy              # Proxy support
├─reader             # RSS, Atom, JSON feed parsers  
├─storage            # Database and file system storage
├─systemd            # Systemd unit files
├─template           # HTML template renderer
├─tests              # Tests
├─timer              # Timer for scheduling  
├─timezone           # Timezone helpers
├─ui                 # Web interface
├─url                # URL parsing
├─validator          # Data validation
├─version            # Version info
└─worker             # Background workers