import styled from 'styled-components'

export const StyledContainer = styled.div`
	position: relative;
	width: 100vw;

	overflow: scroll;
	padding: 50px;
	/* 
    Used for background image. Fills entire Container div. Before used so opacity does not impact children elements */
	&:before {
		display: block;
		content: '';
		height: 100%;
		width: 100%;
		opacity: 0.6;
		position: absolute;
		top: 0;
		left: 0;
	}
`
