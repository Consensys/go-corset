;;error:6:17-26:void expression not permitted here
(defpurefun ((vanishes! :𝔽@loob :force) x) x)
(defcolumns X Y)

(defconstraint c1 ()
  (vanishes! (- (debug X) Y)))
