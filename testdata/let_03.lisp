(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns (X :i16@loob) (Y :i16@loob) (Z :i16))
(defconstraint test ()
  (let ((THREE 3))
    (if X
        (vanishes! 0)
        (vanishes! (- Z (if Y THREE 16))))))
