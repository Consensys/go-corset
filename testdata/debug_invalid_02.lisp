;;error:6:22-31:void expression not permitted here
(defpurefun ((vanishes! :@loob :force) e0) e0)
(defcolumns X Y)

(defconstraint c1 ()
  (vanishes! (- X (* (debug X) Y))))
