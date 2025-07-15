up-rinha:
	@echo "Running up-rinha..."
	@docker-compose -f ./docker/rinha/docker-compose.yml up

up-d-rinha:
	@echo "Running up-rinha on detached mode..."
	@docker-compose -f ./docker/rinha/docker-compose.yml up -d
	@echo "Rinha is running in detached mode."

down-rinha:
	@echo "Stopping up-rinha..."
	@docker-compose -f ./docker/rinha/docker-compose.yml down
	@echo "Rinha stopped."
