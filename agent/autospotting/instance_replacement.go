package autospotting

import "sync"

type instanceReplacement struct {
	wg sync.WaitGroup
}

type message struct {
	eventType        string // example: ping
	instanceID       string
	instanceType     string
	autoScalingGroup string
	// To extend further if needed
}

/*
 This function gets called whenever the Lambda function is invoked as a result of an SNS notification,
 like for example by the cron caller's periodic ping messages
	  - these are sent by our cron lambda function
	  - trigger actions across all ASGs from all regions which were tagged with 'spot-enabled=true'
		- if we have a spot instance that needs to be added to an ASG (uptime is more than the ASG grace period)
			- add it to the ASG
			- detach and terminate an on-demand instance from the biggest AZ, if it contains any on-demand instances
			- If all AZs are equal, pick the one with the most on-demand instances
		- if we have no spot instances to be added, and at least one on-demand instance left in the ASG
			- launch a new spot instance in the AZ where we have the smallest number of instances
				- TODO: in all AZs are equal, we pick the one with the most on-demand instances
			- if other spot instances exist, the new one should have a different type
			- tags and launch configuration are copied from one of the on-demand instances from the ASG
			- launch configuration needs a few changes
				-	instance type - compatible to the original one (HVM/PV, #CPUs, at least as much RAM)
				- user_data is also injected our 'hello' message generator
*/

func (i *instanceReplacement) processAllRegions(instData *jsonInstances) {
	// for each region in parallel
	// for each of the ASGs tagged with 'spot-enabled=true'

	regions, err := getRegions()

	if err != nil {
		logger.Println(err.Error())
		return
	}
	for _, r := range regions {
		i.wg.Add(1)
		r := region{name: r}
		go func() {
			r.processRegion(instData)
			i.wg.Done()
		}()
	}
	i.wg.Wait()

}