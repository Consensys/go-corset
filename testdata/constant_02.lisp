(defcolumns X Y)
;; X*2 == Y*2
(defconstraint c1 () (- (* X (* 2 1)) (* Y (* 1 2))))
;; X*1458 == Y*1458
(defconstraint c1 () (- (* X (* 243 22)) (* Y (* 6 891))))
;; X*2916 == Y*2916
(defconstraint c1 () (- (* X (* 2 243 22)) (* Y (* 6 891 2))))
;; X*2916 == Y*2916
(defconstraint c1 () (- (* X (* 243 2 22)) (* Y (* 6 891 2))))
;; X*2916 == Y*2916
(defconstraint c1 () (- (* X (* 22 243 2)) (* Y (* 6 891 2))))
;; X*2916 == Y*2916
(defconstraint c1 () (- (* X (* 2 243 22)) (* Y (* 891 6 2))))
;; X*2916 == Y*2916
(defconstraint c1 () (- (* X (* 2 243 22)) (* Y (* 2 891 6))))
;; X*2916 == Y*2916
(defconstraint c1 () (- (* X (* 2 243 22)) (* Y (* 2 891 6 1))))
;; X*2916 == Y*2916
(defconstraint c1 () (- (* X (* 2 243 22)) (* Y (* 2 891 6 1 1))))
;; X*2916 == Y*2916
(defconstraint c1 () (- (* X (* 2 243 22 1)) (* Y (* 2 891 6))))
;; X*2916 == Y*2916
(defconstraint c1 () (- (* X (* 2 243 22 1 1)) (* Y (* 2 891 6))))
