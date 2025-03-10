;;error:7:9-10:expected list
;;error:7:11-12:expected list
(defpurefun ((vanishes! :𝔽@loob) x) x)
(defcolumns (A :i16@loob) (B :i16))

(defconstraint c1 ()
  (let (C B)
    (if A
        (vanishes! C))))
