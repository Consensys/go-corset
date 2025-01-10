;;error:6:8-9:expected list
(defpurefun ((vanishes! :@loob) x) x)
(defcolumns (A :@loob) B)

(defconstraint c1 ()
  (let C
    (if A
        (vanishes! C))))
