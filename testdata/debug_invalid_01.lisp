;;error:6:17-26:void expression not permitted here
(defpurefun ((vanishes! :𝔽 :force) x) x)
(defcolumns (X :i16) (Y :i16))

(defconstraint c1 ()
  (vanishes! (- (debug X) Y)))
