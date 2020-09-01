import React from 'react'
import { useState, FC, useEffect } from 'react'
import { makeStyles } from '@material-ui/core/styles'

import Slide from '@material-ui/core/Slide';

import Divider from '@material-ui/core/Divider'
import InputBase from '@material-ui/core/InputBase'
import Paper from '@material-ui/core/Paper'

import AccountCircleIcon from '@material-ui/icons/AccountCircle'
import GpsFixedIcon from '@material-ui/icons/GpsFixed'
import GpsNotFixedIcon from '@material-ui/icons/GpsNotFixed'
import GpsOffIcon from '@material-ui/icons/GpsOff'
import IconButton from '@material-ui/core/IconButton'
import SearchIcon from '@material-ui/icons/Search'

import type { IconButtonProps } from '@material-ui/core/IconButton'

const searchbarCSS = makeStyles((theme) => ({
    searchbar: {
        position: 'relative',
        padding: '2px 4px',
        display: 'flex',
        alignItems: 'center',
        width: '95%',
        margin: '0 auto',
        [theme.breakpoints.down('sm')]: {
            marginTop: '5px',
        },
        [theme.breakpoints.up('sm')]: {
            marginTop: '15px',
        },
        maxWidth: '800px',
        zIndex: 1000,
    },
    input: {
        marginLeft: theme.spacing(1),
        flex: 1,
    },
    iconButton: {
        padding: 10,
    },
    divider: {
        height: 28,
        margin: 4,
    },
}))

type Props = { onGPSTrack: PositionCallback, shown: boolean }
type GPSProps = { status: GPSStatus } & IconButtonProps

enum GPSStatus {
    Disabled,
    Tracking,
    Inactive,
}

const GPSButton: FC<GPSProps> = ({ status, ...props }: GPSProps) => {
    switch (status) {
        case GPSStatus.Inactive:
            return (
                <IconButton {...props}>
                    <GpsNotFixedIcon />
                </IconButton>
            )
        case GPSStatus.Tracking:
            return (
                <IconButton {...props} color="primary">
                    <GpsFixedIcon />
                </IconButton>
            )
        case GPSStatus.Disabled:
            return (
                <IconButton {...props} color="secondary">
                    <GpsOffIcon />
                </IconButton>
            )
    }
}

const Searchbar: FC<Props> = (props: Props) => {
    const { onGPSTrack } = props,
        classes = searchbarCSS(),
        [watchID, setWatchID] = useState(0),
        [gpsStatus, setGPSStatus] = useState(GPSStatus.Inactive),
        gps = navigator.geolocation

    const stopTracking = () => {
        gps.clearWatch(watchID)
        setGPSStatus(GPSStatus.Inactive)
    }

    const watchGPS = () => {
        const opts = { enableHighAccuracy: true }
        setWatchID(
            gps.watchPosition(
                (coords) => {
                    setGPSStatus(GPSStatus.Tracking)
                    onGPSTrack(coords)
                },
                () => {
                    setGPSStatus(GPSStatus.Disabled)
                },
                opts,
            ),
        )
    }

    return (
        <Slide direction="down" in={props.shown} mountOnEnter unmountOnExit>
            <Paper component="form" className={classes.searchbar} elevation={5}>
                <IconButton className={classes.iconButton}>
                    <AccountCircleIcon />
                </IconButton>
                <Divider className={classes.divider} orientation="vertical" />
                <InputBase className={classes.input} placeholder="Sök" />
                <IconButton type="submit" className={classes.iconButton}>
                    <SearchIcon />
                </IconButton>
                <Divider className={classes.divider} orientation="vertical" />
                <GPSButton
                    status={gps ? gpsStatus : GPSStatus.Disabled}
                    className={classes.iconButton}
                    onClick={
                        gpsStatus == GPSStatus.Tracking ? stopTracking : watchGPS
                    }
                />
            </Paper>
        </Slide>
    )
}

export default Searchbar
