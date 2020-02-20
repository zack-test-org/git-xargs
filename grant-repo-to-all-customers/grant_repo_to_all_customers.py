import os
import click
import github
import logging

logging.basicConfig(format='%(asctime)s [%(levelname)s] %(message)s', level=logging.INFO)


@click.command()
@click.option(
    '--repo',
    prompt='Enter the name of the new repo in the gruntwork-io organization to grant teams access to',
    help='Name of the new repo in the gruntwork-io organization to grant teams access to.',
)
@click.option(
    '--seed',
    prompt='Enter the name of a repo in the Gruntwork organization to use as a reference for teams to grant access to',
    help='Name of a repo in the Gruntwork organization to use as a reference for teams to grant access to.',
)
@click.option(
    '--dry/--apply',
    default=True,
    help='Do a dry run, showing all the teams that will be granted access to the repo but not actually grant acces.',
)
def main(repo, seed, dry):
    """
    Grant access to a new repo to all customers who already have access to another repo in the library.
    """
    token = os.environ.get('GITHUB_OAUTH_TOKEN')
    if not token:
        raise click.ClickException('Environment variable GITHUB_OAUTH_TOKEN is not set')

    clt = github.GithubApiClient(token)

    logging.info(f'Retrieving all teams that have access to repo {seed}')
    teams = github.get_repository_teams(clt, seed)
    customer_teams = [t for t in teams if t.startswith('client-')]
    logging.info(f'Found {len(customer_teams)} customer teams that have access to {seed}')

    logging.info('Granting all found teams access to repo')
    for team in customer_teams:
        logging.info(f'Granting access to {repo} to {team}')
        if dry:
            logging.info('DRY RUN: Skipping')
        else:
            github.add_repo_to_team(clt, team, repo)
            logging.info('Successfully granted access')


if __name__ == '__main__':
    main()
