;;error:7:9-10:expected list
;;error:7:11-12:expected list
(defpurefun ((vanishes! :@loob) x) x)
(defcolumns (A :@loob) B)

(defconstraint c1 ()
  (let (C B)
    (if A
        (vanishes! C))))
