;;error:6:22-31:void expression not permitted here
(defpurefun ((vanishes! :𝔽@loob :force) x) x)
(defcolumns X Y)

(defconstraint c1 ()
  (vanishes! (- X (* (debug X) Y))))
